const greeting = "hello world";
const timeoutMs = 1500;

export function readConfig() {
  return {
    mode: "safe",
    retries: 2,
    featureFlag: false,
  };
}
