import test from 'node:test';
import assert from 'node:assert/strict';
import {readFileSync} from 'node:fs';
import {TransactionBuilder,Account,Operation,Asset,xdr} from '@stellar/stellar-sdk';
import {verify} from './verify.mjs';
const live=JSON.parse(readFileSync(new URL('./testnet.json',import.meta.url)));
const copy=()=>structuredClone(live);
test('captured public testnet envelope/result/meta/events agree with Horizon',()=>{
 const r=verify(live);assert.equal(r.status,'EQUIVALENT');assert.equal(r.payments.length,live.horizon.length);assert.equal(r.coverage_complete,false);
});
test('duplicate observations do not create extra settlements; conflicting duplicates fail',()=>{
 const f=copy();f.horizon.push(structuredClone(f.horizon[0]));assert.equal(verify(f).payments.length,live.horizon.length);
 f.horizon.at(-1).amount='1';assert.throws(()=>verify(f));
});
test('wrong network, amount, issuer, direction and operation identity fail',()=>{
 for(const edit of [f=>f.network='other network',f=>f.horizon[0].amount='1',f=>f.horizon[0].asset_issuer='other',f=>f.horizon[0].from=f.horizon[0].to,f=>f.horizon[0].id='1']) {const f=copy();edit(f);assert.throws(()=>verify(f));}
});
test('NOT_FOUND and retention gap cannot prove missing settlement',()=>{
 const f=copy();f.rpc.status='NOT_FOUND';assert.equal(verify(f).status,'UNKNOWN');
 const g=copy();g.rpc.oldestLedger=g.rpc.ledger+1;assert.equal(verify(g).status,'UNKNOWN');
});
test('failed transaction and contradictory status',()=>{
 const f=copy();f.rpc.status='FAILED';assert.throws(()=>verify(f));
 const r=xdr.TransactionResult.fromXDR(f.rpc.resultXdr,'base64');r.result=xdr.TransactionResultResult.txBadSeq();f.rpc.resultXdr=r.toXDR('base64');assert.equal(verify(f).status,'UNKNOWN');
});
test('missing and duplicate classic events do not become settlements',()=>{
 const f=copy();const meta=xdr.TransactionMeta.fromXDR(f.rpc.resultMetaXdr,'base64');meta.v4.operations[0].events=[];f.rpc.resultMetaXdr=meta.toXDR('base64');delete f.rpc.events;assert.equal(verify(f).status,'UNKNOWN');
 const g=copy();g.rpc.events.contractEventsXdr[0].push(g.rpc.events.contractEventsXdr[0][0]);assert.throws(()=>verify(g));
});
test('synthetic two-payment batch maps distinct operation IDs without event double counting',()=>{
 const f=copy();f.evidence='synthetic mutation; never submitted';const h=f.horizon[0];
 const tx=new TransactionBuilder(new Account(h.from,'1'),{fee:'100',networkPassphrase:f.network}).addOperation(Operation.payment({destination:h.to,asset:Asset.native(),amount:h.amount})).addOperation(Operation.payment({destination:h.to,asset:Asset.native(),amount:h.amount})).setTimeout(0).build();
 f.rpc.envelopeXdr=tx.toXDR();f.rpc.txHash=Buffer.from(tx.hash()).toString('hex');
 const r=xdr.TransactionResult.fromXDR(f.rpc.resultXdr,'base64');r.result=xdr.TransactionResultResult.txSuccess([r.result.results[0],r.result.results[0]]);f.rpc.resultXdr=r.toXDR('base64');
 const m=xdr.TransactionMeta.fromXDR(f.rpc.resultMetaXdr,'base64');m.v4.operations=[m.v4.operations[0],m.v4.operations[0]];f.rpc.resultMetaXdr=m.toXDR('base64');f.rpc.events.contractEventsXdr=[f.rpc.events.contractEventsXdr[0],f.rpc.events.contractEventsXdr[0]];
 f.horizon=[{...h,transaction_hash:f.rpc.txHash},{...h,transaction_hash:f.rpc.txHash,id:(BigInt(h.id)+1n).toString()}];
 const checked=verify(f);assert.equal(checked.status,'EQUIVALENT');assert.equal(checked.payments.length,2);assert.notEqual(checked.payments[0].operation_id,checked.payments[1].operation_id);
});
