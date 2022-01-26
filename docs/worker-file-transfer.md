# wharf-cmd-worker file transfer

Per Wharf build stage:

1. Spin up pod with:

   - init container:

     - image: `alpine:3`
     - command: `sh -c 'sleep infinite || true'`
     - volume mounts: `repo` at `/mnt/repo`

   - app container: whatever the build step needs

     - volume mounts: `repo` at `/mnt/repo`
   
   - volume:
   
     - `repo` as an `emptyDir: {}`

2. Transfer repo over to alpine init container, running the equivalence of:

   ```sh
   kubectl cp ./ podname:/mnt/repo --container init
   ```

3. Execute `killall -s SIGINT sleep` on init container to make it stop sleeping.

4. App container starts running

## Example

```sh
kubectl apply -f ./worker-file-transfer-pod.yaml

# wait for status being "Init:0/1"

kubectl cp ./ wharf-build-step-test:/mnt/repo -c init

kubectl exec wharf-build-step-test -c init -- killall -s SIGINT sleep

kubectl logs wharf-build-step-test
```
