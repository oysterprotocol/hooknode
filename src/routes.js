import * as iotaController from "./controllers/iota-controller";

export default app => {
  app.post("/HookListener.php", iotaController.legacy);

  // TODO: Mirror IRI API.

  return app;
};
