import run from "./jsonrpc"

import link, { linkMethod } from "./lib/link";
import start, { startMethod } from "./lib/start";

import { ensureChainSetup, ensureChainSetupMethod } from "./lib/chain";
import { connectChains, connectChainsMethod } from "./lib/chain";
import { getPath, getPathMethod } from "./lib/chain";
import { listPaths, listPathsMethod } from "./lib/chain";
import { getDefaultAccount, getDefaultAccountMethod } from "./lib/chain";
import { getDefaultAccountBalance, getDefaultAccountBalanceMethod } from "./lib/chain";

run([
  [linkMethod, link],
  [startMethod, start],
  [ensureChainSetupMethod, ensureChainSetup],
  [connectChainsMethod, connectChains],
  [getPathMethod, getPath],
  [listPathsMethod, listPaths],
  [getDefaultAccountMethod, getDefaultAccount],
  [getDefaultAccountBalanceMethod, getDefaultAccountBalance],
]);
