# ASM Webhook
![Build](https://github.com/ayoul3/asm-webhook/workflows/Go/badge.svg)

asm-webhook is a mutating Webhook designed to dynamically fetch secrets from SecretsManager and inject them as env varibales in pods.

It is heavily inspired from the great [Banzai Vaults](https://github.com/banzaicloud/bank-vaults/tree/master/charts/vault-secrets-webhook) that only supports Vault.

## How does it work?
1. Interception

A mutating webhook will intercept all pod submissions bearing the annotation `asm-webhook:"true"`.

2. Validation

It will go through all its environment variables and secrets looking for SecretsManager ARNs. If found, the Pod will be mutated.

3. Mutation

An init container `ayoul3/asm-env` is injected which copies its main binary `asm-env` to a volume `/asm` shared with the other containers in the pod.

4. Execution

The command of the container is overwritten so that its fist starts a binary `asm-env` that will decrypt SecretsManager secrets. After which it will start the original command with its arguments.


```yaml
Command: sh
Args:
 -c
 - trap 'exit' TERM; while :; do sleep 1; echo decrypted $KEY_ID; done
```
becomes:
```yaml
Command: /asm/asm-env
Args:
 - sh
 -c
 - trap 'exit' TERM; while :; do sleep 1; echo decrypted $KEY_ID; done
```

## Install
Default values work just fine if you want to test on minikube.

You can always change values in `./chart/values.yaml` to match your naming convention, namespace, labels and so on.

Once you're done execute the `generate.sh` script to provision certificates used by the webhook:

```bash
./generate.sh
```
This will gnerate a `secret` resource in Kube that will be mounted and used by the Webhook.

If everything goes alright, you can then deploy the chart using helm:
```bash
helm upgrade --install asm-webhook ./chart
```

Check that the webhook is running:
```bash
$ kubectl get mutatingwebhookconfigurations

NAME                      WEBHOOKS   AGE
asm-webhook.default.svc   1          17s

$ k get deployments

NAME          READY   UP-TO-DATE   AVAILABLE   AGE
asm-webhook   1/1     1            1           17s
...snip...
```

## Prerequisites

It is the pod that fetches its own secrets, so obviously it needs to use a service account mapped to an IAM role capable of reading such secrets. You can read more about it [here](https://docs.aws.amazon.com/eks/latest/userguide/create-service-account-iam-policy-and-role.html) and find an actual example [here](https://aws.amazon.com/blogs/containers/aws-secrets-controller-poc/).

## Secret formats
Secrets stored in SecretsManager can be of two formats:
* Simple strings
* Flat JSON
You can specify which JSON key to fetch by adding it after the character **#**.

If the secret `arn:aws:secretsmanager:us-east-1:886477354405:secret:key-us-cmo1Hc` contains  `{"user":"test", "password": "secret"}` then you can choose to only fetch the password key as such:
```yaml
DB_PASS: arn:aws:secretsmanager:us-east-1:886477354405:secret:key-us-cmo1Hc#password
```
*Nested keys are not supported at the moment.*

## Annotations

The webhook supports the following annotations at the pod level:
| Tables                              | Decription    | Value    |
| ----------------------------------- |:-------------:| :----------:|
| asm.webhook.debug                   | Activate debug logs for the webhook handler of this pod | false/true
| asm.webhook.asm-env.image           | image of the init container to inject      | `ayoul3/asm-env:latest`
| asm.webhook.asm-env.path            | path to binary inside the init container that will fetch secrets     | /app/
| asm.webhook.asm-env.bin             | name of the binary that will fetch secrets       | asm-env
| asm.webhook.asm-env.mountPath       | mount path where containers will find this binary      | /asm/
| asm.webhook.asm-env.skip-cert-check | skip certificate check when contacting SecretsManager. Useful for bare pods without `ca-certificates` package      | false

## asm-env

asm-env binary is maintained as a [separate project](https://github.com/ayoul3/asm-env).

## Monitoring

The webhook exposes Prometheus metrics on `/metrics`. A [grafana dashboard](https://grafana.com/grafana/dashboards/13685) is also available courtesy of Xabier Larrakoetxea.


## Credit
[Banzai Vaults](https://github.com/banzaicloud/bank-vaults/tree/master/charts/vault-secrets-webhook)
[Kubewebhook](https://github.com/slok/kubewebhook)