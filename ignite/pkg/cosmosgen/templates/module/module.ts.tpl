// Generated by Ignite ignite.com/cli

import { StdFee } from "@cosmjs/launchpad";
import { SigningStargateClient, DeliverTxResponse } from "@cosmjs/stargate";
import { EncodeObject, GeneratedType, OfflineSigner, Registry } from "@cosmjs/proto-signing";
import { msgTypes } from './registry';
import { IgniteClient } from "../client"
import { MissingWalletError } from "../helpers"
import { Api } from "./rest";
{{ range .Module.Msgs }}import { {{ .Name }} } from "./types/{{ resolveFile .FilePath }}";
{{ end }}

export { {{ range $i,$type:=.Module.Types }}{{ if (gt $i 0) }}, {{ end }}{{ $type.Name }}{{ end }} };
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

const defaultFee = {
  amount: [],
  gas: "200000",
};

class SDKModule extends Api<any> {
	private _signer: OfflineSigner;
	private _rpcAddr: string;
	private _prefix: string;
	public registry: Array<[string, GeneratedType]>;

	constructor(client: IgniteClient) {		
		super({baseUrl: client.env.apiURL});
		this._signer = client.env.signer;		
		this._rpcAddr = client.env.rpcURL;
		this._prefix = client.env.prefix ?? 'cosmos';
	}


	{{ range .Module.Msgs }}
	async send{{ .Name }}({ value, fee, memo }: send{{ .Name }}Params): Promise<DeliverTxResponse> {
		if (!this._signer) {
		    throw new Error('TxClient:send{{ .Name }}: Unable to sign Tx. Signer is not present.')
		}
		if (!this._rpcAddr) {
            throw new Error('TxClient:send{{ .Name }}: Unable to sign Tx. Address is not present.')
        }
		try {
			const signingClient = await SigningStargateClient.connectWithSigner(this._rpcAddr,this._signer,{registry: new Registry(this.registry), prefix:this._prefix});
			let msg = this.{{ camelCase .Name }}({ value: {{ .Name }}.fromPartial(value) })
			return await signingClient.signAndBroadcast(this._rpcAddr, [msg], fee ? fee : { amount: [], gas: '200000' }, memo)
		} catch (e: any) {
			throw new Error('TxClient:send{{ .Name }}: Could not broadcast Tx: '+ e.message)
		}
	}
	{{ end }}
	{{ range .Module.Msgs }}
	{{ camelCase .Name }}({ value }: {{ camelCase .Name }}Params): EncodeObject {
		try {
			return { typeUrl: "/{{ .URI }}", value: {{ .Name }}.fromPartial( value ) }  
		} catch (e: any) {
			throw new Error('TxClient:{{ .Name }}: Could not create message: ' + e.message)
		}
	}
	{{ end }}
};

const Module = (test: IgniteClient) => {
	return {
		module: {
			{{ camelCaseUpperSta .Module.Pkg.Name }}: new SDKModule(test)
		},
		registry: msgTypes
  }
}
export default Module;