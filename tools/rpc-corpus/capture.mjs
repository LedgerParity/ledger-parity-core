// Opt-in public-testnet reads only. JSON-RPC POST invokes read methods;
// this script never signs, funds, submits, or simulates a transaction.
import { writeFile } from 'node:fs/promises';
const horizon = 'https://horizon-testnet.stellar.org';
const rpc = 'https://soroban-testnet.stellar.org';
const network = 'Test SDF Network ; September 2015';
async function read(url, options = {}) {
  const response = await fetch(url, {...options, redirect:'error',signal:AbortSignal.timeout(20000)});
  if (!response.ok) throw new Error(`HTTP ${response.status}`);
  return response.json();
}
async function call(method, params={}) {
  const r=await read(rpc,{method:'POST',headers:{'Content-Type':'application/json'},body:JSON.stringify({jsonrpc:'2.0',id:1,method,params})});
  if(r.error) throw new Error(JSON.stringify(r.error));return r.result;
}
if((await call('getNetwork')).passphrase!==network) throw new Error('Unexpected RPC network');
if((await read(horizon)).network_passphrase!==network) throw new Error('Unexpected Horizon network');
const recent=(await read(`${horizon}/payments?order=desc&limit=100`))._embedded.records;
let captured;
for(const payment of recent.filter(p=>p.type==='payment'&&p.transaction_successful)) {
  const result=await call('getTransaction',{hash:payment.transaction_hash});
  if(result.status!=='SUCCESS') continue;
  const operations=(await read(`${horizon}/transactions/${payment.transaction_hash}/operations?limit=200&order=asc`))._embedded.records;
  // The first corpus targets ordinary-payment-only transactions, not mixed
  // contract/path/account-creation semantics. No silent operation omission.
  if(!operations.length||operations.some(p=>p.type!=='payment')) continue;
  captured={schema:'ledgerparity-rpc-corpus/v1',evidence:'public-testnet provider observations; not independent application records',captured_at:new Date().toISOString(),network,horizon_url:horizon,rpc_url:rpc,rpc:result,horizon:operations};
  break;
}
if(!captured) throw new Error('No supported recent ordinary-payment transaction; no fixture written');
const path=process.argv[2];if(!path)throw new Error('Supply a new output filename');
await writeFile(path,JSON.stringify(captured,null,2)+'\n',{flag:'wx',mode:0o600});
console.log(`Captured ${captured.horizon.length} ordinary payments to ${path}; verification required`);
