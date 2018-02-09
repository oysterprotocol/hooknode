import IOTA from "iota.lib.js";

const CONFIG = Object.freeze({
  PROVIDER: new IOTA({ provider: "http://localhost:14265" }),
  MIN_DEPTH: 1,
  MIN_WEIGHT_MAGNITUDE: 14,
});

export const legacy = (req, res) => {
  // TODO: Verify broker middleware.

  const { trytes } = req.body;
  const { PROVIDER, MIN_DEPTH, MIN_WEIGHT_MAGNITUDE } = CONFIG;

  console.log("BODY");
  console.log(req.body);

  PROVIDER.api.sendTrytes(trytes, MIN_DEPTH, MIN_WEIGHT_MAGNITUDE);

  // Async response
  return res.status(204).send("success");
};
