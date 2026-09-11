import assert from 'node:assert/strict';
import {TransactionBuilder,xdr,scValToNative,StrKey} from '@stellar/stellar-sdk';

function units(amount) {
  assert.match(amount,/^\d+(\.\d{1,7})?$/);
  const [whole,fraction='']=amount.split('.');const n=BigInt(whole)*10000000n+BigInt(fraction.padEnd(7,'0'));
  assert(n>0n&&n<=9223372036854775807n);return n;
}
const unknown=reason=>({status:'UNKNOWN',reason,coverage_complete:false,payments:[]});

// Deliberately a corpus verifier, not an account-history ingestor. It compares
// envelope intent + successful operation results + operation-scoped asset events
// with a second provider's ordinary-payment observations.
export function verify(f) {
  assert.equal(f.schema,'ledgerparity-rpc-corpus/v1');assert(f.network);
  const r=f.rpc;
  if(r.status==='NOT_FOUND') return unknown('Transaction lookup cannot prove absence or history coverage');
  assert(['SUCCESS','FAILED'].includes(r.status));
  const result=xdr.TransactionResult.fromXDR(r.resultXdr,'base64');
  if(r.status==='FAILED') {assert.notEqual(result.result.type,'txSuccess');return unknown('Failed transaction is not settlement');}
  if(r.feeBump) return unknown('Fee-bump envelopes are outside this corpus verifier');
  assert.equal(result.result.type,'txSuccess');
  assert(Number.isInteger(r.ledger)&&r.ledger>0&&r.ledger<=2147483647);
  assert(Number.isInteger(r.applicationOrder)&&r.applicationOrder>0&&r.applicationOrder<1048576);
  if(r.ledger<r.oldestLedger||r.ledger>r.latestLedger) return unknown('Transaction outside declared retained ledger range');
  const tx=TransactionBuilder.fromXDR(r.envelopeXdr,f.network);
  const hash=Buffer.from(tx.hash()).toString('hex');assert.equal(hash,r.txHash);
  const meta=xdr.TransactionMeta.fromXDR(r.resultMetaXdr,'base64');
  if(meta.type!=='v4') return unknown('Only V4 metadata with operation-scoped classic events is verified');
  const operations=tx.operations;assert(operations.length>0&&operations.length<=100);
  if(operations.some(o=>o.type!=='payment')) return unknown('Mixed/path/contract operations are unsupported');
  assert.equal(meta.v4.operations.length,operations.length);assert.equal(result.result.results.length,operations.length);
  const timestamp=new Date(Number(r.createdAt)*1000).toISOString();
  assert(Number.isSafeInteger(Number(r.createdAt))&&Number(r.createdAt)>0);
  const payments=[];
  for(let i=0;i<operations.length;i++) {
    const o=operations[i],source=o.source||tx.source;
    if(!source.startsWith('G')||!o.destination.startsWith('G'))return unknown('Muxed account identity needs separate fixtures');
    const res=result.result.results[i];assert.equal(res.type,'opInner');assert.equal(res.tr.type,'payment');assert.equal(res.tr.paymentResult.type,'paymentSuccess');
    const events=meta.v4.operations[i].events;
    if(!events.length)return unknown('Classic asset events unavailable; no equivalence claim');
    // The separate RPC events field mirrors metadata, not extra settlements.
    if(r.events?.contractEventsXdr) assert.deepEqual(r.events.contractEventsXdr[i],events.map(e=>e.toXDR('base64')));
    assert.equal(events.length,1,'Ambiguous operation event count');
    const e=events[0];assert.equal(e.type.name,'contract');assert.equal(e.body.type,'v0');
    const topics=e.body.v0.topics.map(scValToNative);const native=o.asset.isNative();
    assert.deepEqual(topics,['transfer',source,o.destination,native?'native':`${o.asset.getCode()}:${o.asset.getIssuer()}`]);
    assert.equal(StrKey.encodeContract(e.contractId.value),o.asset.contractId(f.network));
    const data=scValToNative(e.body.v0.data);assert.equal(typeof data==='bigint'?data:data.amount,units(o.amount));
    // Event memo/muxed metadata is not substituted for envelope account identity.
    const id=((BigInt(r.ledger)<<32n)|(BigInt(r.applicationOrder)<<12n)|BigInt(i+1)).toString();
    payments.push({operation_id:id,transaction_hash:hash,account:source,destination:o.destination,amount_units:units(o.amount).toString(),asset_type:o.asset.getAssetType(),asset_code:o.asset.getCode(),asset_issuer:o.asset.getIssuer()||'',timestamp});
  }
  const seen=new Map();
  for(const h of f.horizon) {
    assert.equal(h.type,'payment');assert.equal(h.transaction_successful,true);assert.equal(h.transaction_hash,hash);
    if(h.from_muxed||h.to_muxed)return unknown('Muxed Horizon account requires separate fixtures');
    const p={operation_id:h.id,transaction_hash:h.transaction_hash,account:h.from,destination:h.to,amount_units:units(h.amount).toString(),asset_type:h.asset_type,asset_code:h.asset_type==='native'?'XLM':h.asset_code,asset_issuer:h.asset_issuer||'',timestamp:new Date(h.created_at).toISOString()};
    if(seen.has(h.id))assert.deepEqual(seen.get(h.id),p,'Conflicting duplicate Horizon observation');
    seen.set(h.id,p);
  }
  assert.deepEqual([...seen.values()].sort((a,b)=>a.operation_id.localeCompare(b.operation_id)),[...payments].sort((a,b)=>a.operation_id.localeCompare(b.operation_id)));
  return {status:'EQUIVALENT',coverage_complete:false,payments};
}
