import IOTA from "iota.lib.js";

const CONFIG = Object.freeze({
  PROVIDER: new IOTA({ provider: "http://localhost:14265" }),
  MIN_DEPTH: 3,
  MIN_WEIGHT_MAGNITUDE: 14,
});

export const legacy = (req, res) => {
  // TODO: Verify broker middleware.

  const { trytes } = req.body;
  const { PROVIDER, MIN_DEPTH, MIN_WEIGHT_MAGNITUDE } = CONFIG;

  PROVIDER.api.sendTrytes(trytes, MIN_DEPTH, MIN_WEIGHT_MAGNITUDE, console.log);

  // Async response
  return res.status(204).send("success");
};

export const debug = (req, res) => {
  console.log("testing123...");

  const { PROVIDER } = CONFIG;
  PROVIDER.api.findTransactionObjects(
    {
      addresses: [
        "9BUABBUAPCRCVAZARCBBPCPCRCBBWAPCPCWAUCBBXAPCWAZA9BBBYASC9BQCXAXAWAQCWAYABBWAUCUAUDWZNUMLSB",
      ],
    },
    console.log,
  );

  res.status(204).send("success");
};
