package typed

import (
	"fmt"
	"strings"

	"github.com/ignite/cli/v29/ignite/pkg/multiformatname"
	"github.com/ignite/cli/v29/ignite/pkg/placeholder"
	"github.com/ignite/cli/v29/ignite/pkg/xast"
)

func ModuleSimulationMsgModify(
	content string,
	_, _ multiformatname.Name,
	_ ...string,
) (string, error) {
	if len(msgs) == 0 {
		msgs = append(msgs, "")
	}

	var err error
	for _, msg := range msgs {
		// simulation operations
		replacementOp := fmt.Sprintf(
			`reg.Add(weights.Get("msg_%[3]v", 100 /* determine the simulation weight value */), simulation.Msg%[1]v%[2]vFactory(am.keeper))`,
			msg,
			typeName.UpperCamel,
			fmt.Sprintf("%s_%s", strings.ToLower(msg), typeName.Snake),
		)
		content, err = xast.ModifyFunction(content, "WeightedOperations", xast.AppendFuncCode(replacementOp))
		if err != nil {
			return "", err
		}

		// add proposal simulation operations for msgs having an authority as signer.
		if strings.Contains(content, "ProposalMsgs") && strings.EqualFold(msgSigner.Original, "authority") {
			replacementOpMsg := fmt.Sprintf(
				`reg.Add(weights.Get("msg_%[2]v", 100), simulation.Msg%[1]v%[2]vFactory(am.keeper))`,
				msg,
				typeName.UpperCamel,
				fmt.Sprintf("%s_%s", strings.ToLower(msg), typeName.Snake),
			)
			content, err = xast.ModifyFunction(content, "ProposalMsgs", xast.AppendFuncCode(replacementOpMsg))
			if err != nil {
				return "", err
			}
		}
	}

	return content, nil
}

func OldModuleSimulationMsgModify(
	replacer placeholder.Replacer,
	content,
	moduleName string,
	typeName multiformatname.Name,
	msgs ...string,
) string {
	var (
		PlaceholderSimappConst        string
		PlaceholderSimappOperation    string
		PlaceholderSimappOperationMsg string
	)

	if len(msgs) == 0 {
		msgs = append(msgs, "")
	}
	for _, msg := range msgs {
		// simulation constants
		templateConst := `opWeightMsg%[2]v%[3]v = "op_weight_msg_%[4]v"
	// TODO: Determine the simulation weight value
	defaultWeightMsg%[2]v%[3]v int = 100

	%[1]v`
		replacementConst := fmt.Sprintf(templateConst, PlaceholderSimappConst, msg, typeName.UpperCamel, typeName.Snake)
		content = replacer.Replace(content, PlaceholderSimappConst, replacementConst)

		// simulation operations
		templateOp := `var weightMsg%[2]v%[3]v int
	simState.AppParams.GetOrGenerate(opWeightMsg%[2]v%[3]v, &weightMsg%[2]v%[3]v, nil,
		func(_ *rand.Rand) {
			weightMsg%[2]v%[3]v = defaultWeightMsg%[2]v%[3]v
		},
	)
	operations = append(operations, simulation.NewWeightedOperation(
		weightMsg%[2]v%[3]v,
		%[4]vsimulation.SimulateMsg%[2]v%[3]v(am.accountKeeper, am.bankKeeper, am.keeper),
	))

	%[1]v`
		replacementOp := fmt.Sprintf(templateOp, PlaceholderSimappOperation, msg, typeName.UpperCamel, moduleName)
		content = replacer.Replace(content, PlaceholderSimappOperation, replacementOp)

		if strings.Contains(content, PlaceholderSimappOperationMsg) {
			templateOpMsg := `simulation.NewWeightedProposalMsg(
	opWeightMsg%[2]v%[3]v,
	defaultWeightMsg%[2]v%[3]v,
	func(r *rand.Rand, ctx sdk.Context, accs []simtypes.Account) sdk.Msg {
		%[4]vsimulation.SimulateMsg%[2]v%[3]v(am.accountKeeper, am.bankKeeper, am.keeper)
		return nil
	},
),
%[1]v`
			replacementOpMsg := fmt.Sprintf(templateOpMsg, PlaceholderSimappOperationMsg, msg, typeName.UpperCamel, moduleName)
			content = replacer.Replace(content, PlaceholderSimappOperationMsg, replacementOpMsg)
		}
	}
	return content
}
