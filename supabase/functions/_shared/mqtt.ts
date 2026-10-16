import mqtt, { type MqttClient } from "npm:mqtt@5.10.1";

// The machine firmware subscribes to this topic; don't change it or the payload format.
const DISPENSE_TOPIC = "chapa/webhook";
const TIMEOUT_MS = 5000;

function withTimeout<T>(p: Promise<T>, what: string): Promise<T> {
  let timer: number;
  const timeout = new Promise<never>((_, reject) => {
    timer = setTimeout(() => reject(new Error(`${what} timed out after ${TIMEOUT_MS}ms`)), TIMEOUT_MS);
  });
  return Promise.race([p, timeout]).finally(() => clearTimeout(timer));
}

export async function connectMQTT(): Promise<MqttClient> {
  const brokerUrl = Deno.env.get("MQTT_BROKER_URL");
  if (!brokerUrl) throw new Error("MQTT_BROKER_URL is not set");

  const client = mqtt.connect(brokerUrl, {
    username: Deno.env.get("MQTT_USERNAME"),
    password: Deno.env.get("MQTT_PASSWORD"),
    connectTimeout: TIMEOUT_MS,
    reconnectPeriod: 0, // serverless: one connection per request
  });
  try {
    await withTimeout(
      new Promise<void>((resolve, reject) => {
        client.once("connect", () => resolve());
        client.once("error", reject);
      }),
      "MQTT connect",
    );
    return client;
  } catch (err) {
    client.end(true);
    throw err;
  }
}

export async function publishDispense(client: MqttClient, motorCodes: string): Promise<void> {
  await withTimeout(client.publishAsync(DISPENSE_TOPIC, motorCodes, { qos: 1 }), "MQTT publish");
  console.log(`[MQTT] published "${motorCodes}" to ${DISPENSE_TOPIC}`);
}

export async function closeMQTT(client: MqttClient): Promise<void> {
  await withTimeout(client.endAsync(), "MQTT close").catch(() => client.end(true));
}

export async function dispense(motorCodes: string): Promise<void> {
  const client = await connectMQTT();
  try {
    await publishDispense(client, motorCodes);
  } finally {
    await closeMQTT(client);
  }
}
