// Generated by Ignite ignite.com/cli

import { SigningStargateClient, DeliverTxResponse, StdFee } from "@cosmjs/stargate";
import { EncodeObject, GeneratedType, OfflineSigner, Registry } from "@cosmjs/proto-signing";
import { msgTypes } from './registry';
import { IgniteClient } from "../client"
import { MissingWalletError } from "../helpers"
import { Api } from "./rest";
{{ range .Module.Msgs }}import { {{ .Name }} } from "./types/{{ resolveFile .FilePath }}";
{{ end }}
{{ range .Module.Types }}import { {{ .Name }} as type{{- .Name -}} } from "./types"
{{ end }}
export { {{ range $i,$type:=.Module.Msgs }}{{ if (gt $i 0) }}, {{ end }}{{ $type.Name }}{{ end }} };
{{ range .Module.Msgs }}
type send{{ .Name }}Params = {
  value: {{ .Name }},
  fee?: StdFee,
  memo?: string
};
{{ end }}
{{ range .Module.Msgs }}
type {{ camelCase .Name }}Params = {
  value: {{ .Name }},
};
{{ end }}

export const registry = new Registry(msgTypes);

type Field = {
	name: string;
	type: unknown;
}
function getStructure(template) {
	const structure: {fields: Field[]} = { fields: [] }
	for (let [key, value] of Object.entries(template)) {
		let field = { name: key, type: typeof value }
		structure.fields.push(field)
	}
	return structure
}
const defaultFee = {
  amount: [],
  gas: "200000",
};

interface TxClientOptions {
  addr: string
	prefix: string
	signer?: OfflineSigner
}

export const txClient = ({ signer, prefix, addr }: TxClientOptions = { addr: "http://localhost:26657", prefix: "cosmos" }) => {

  return {
		{{ range .Module.Msgs }}
		async send{{ .Name }}({ value, fee, memo }: send{{ .Name }}Params): Promise<DeliverTxResponse> {
			if (!signer) {
					throw new Error('TxClient:send{{ .Name }}: Unable to sign Tx. Signer is not present.')
			}
			try {			
				const { address } = (await signer.getAccounts())[0]; 
				const signingClient = await SigningStargateClient.connectWithSigner(addr,signer,{registry});
				let msg = this.{{ camelCase .Name }}({ value: {{ .Name }}.fromPartial(value) })
				return await signingClient.signAndBroadcast(address, [msg], fee ? fee : defaultFee, memo)
			} catch (e: any) {
				throw new Error('TxClient:send{{ .Name }}: Could not broadcast Tx: '+ e.message)
			}
		},
		{{ end }}
		{{ range .Module.Msgs }}
		{{ camelCase .Name }}({ value }: {{ camelCase .Name }}Params): EncodeObject {
			try {
				return { typeUrl: "/{{ .URI }}", value: {{ .Name }}.fromPartial( value ) }  
			} catch (e: any) {
				throw new Error('TxClient:{{ .Name }}: Could not create message: ' + e.message)
			}
		},
		{{ end }}
	}
};

interface QueryClientOptions {
  addr: string
}

export const queryClient = ({ addr: addr }: QueryClientOptions = { addr: "http://localhost:1317" }) => {
  return new Api({ baseURL: addr });
};

class SDKModule {
	public query: ReturnType<typeof queryClient>;
	public tx: ReturnType<typeof txClient>;
	public structure: Record<string,unknown>;
	public registry: Array<[string, GeneratedType]> = [];

	constructor(client: IgniteClient) {		
	
		this.query = queryClient({ addr: client.env.apiURL });		
		this.updateTX(client);
		this.structure =  {
						{{ range .Module.Types }}{{ .Name }}: getStructure(type{{ .Name }}.fromPartial({})),
						{{ end }}
		};
		client.on('signer-changed',(signer) => {			
		 this.updateTX(client);
		})
	}
	updateTX(client: IgniteClient) {
    const methods = txClient({
        signer: client.signer,
        addr: client.env.rpcURL,
        prefix: client.env.prefix ?? "cosmos",
    })
	
    this.tx = methods;
    for (let m in methods) {
        this.tx[m] = methods[m].bind(this.tx);
    }
	}
};

const IgntModule = (test: IgniteClient) => {
	return {
		module: {
			{{ camelCaseUpperSta .Module.Pkg.Name }}: new SDKModule(test)
		},
		registry: msgTypes
  }
}
export default IgntModule;