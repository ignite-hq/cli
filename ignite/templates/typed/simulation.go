package typed

import (
	"github.com/ignite/cli/v29/ignite/pkg/multiformatname"
)

// TODO(@julienrbrt): remove this line when simulation is brought back.
func ModuleSimulationMsgModify(
	content string,
	_, _ multiformatname.Name,
	_ ...string,
) (string, error) {
	// if len(msgs) == 0 {
	// 	msgs = append(msgs, "")
	// }

	// var err error
	// for _, msg := range msgs {
	// 	// simulation operations
	// 	replacementOp := fmt.Sprintf(
	// 		`reg.Add(weights.Get("msg_%[3]v", 100 /* determine the simulation weight value */), simulation.Msg%[1]v%[2]vFactory(am.keeper))`,
	// 		msg,
	// 		typeName.UpperCamel,
	// 		fmt.Sprintf("%s_%s", strings.ToLower(msg), typeName.Snake),
	// 	)
	// 	content, err = xast.ModifyFunction(content, "WeightedOperationsX", xast.AppendFuncCode(replacementOp))
	// 	if err != nil {
	// 		return "", err
	// 	}

	// 	// add proposal simulation operations for msgs having an authority as signer.
	// 	if strings.Contains(content, "ProposalMsgsX") && strings.EqualFold(msgSigner.Original, "authority") {
	// 		replacementOpMsg := fmt.Sprintf(
	// 			`reg.Add(weights.Get("msg_%[2]v", 100), simulation.Msg%[1]v%[2]vFactory(am.keeper))`,
	// 			msg,
	// 			typeName.UpperCamel,
	// 			fmt.Sprintf("%s_%s", strings.ToLower(msg), typeName.Snake),
	// 		)
	// 		content, err = xast.ModifyFunction(content, "ProposalMsgsX", xast.AppendFuncCode(replacementOpMsg))
	// 		if err != nil {
	// 			return "", err
	// 		}
	// 	}
	// }

	return content, nil
}
