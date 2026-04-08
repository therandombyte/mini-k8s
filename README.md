## Learning Roadmap (rewiring the Control plane ideas)

### Phase 1: Single Binary API Server + In-memory store + Pod object only

Building the API Server first as it defines the Object contract, and other components to consume the contract through watches and status updates

### Phase 2: Separate Scheduler and Node Agent (kubelet like), add watch stream, add Pos Status udpate

### Phase 3: Add Simple Deployment Controller and rolling replacement logic

### Phase 4: Add Persistence, auth, basic service abstraction

## Scaffolding

1. cmd: for binaries.
2. cmd/mk-apiserver: load config, initialize store, initialize watches, server HTTP, persist objects.
3. cmd/mkctl: a layer over api server to debug the system faster, like kubectl.
4. cmd/mk-scheduler: watch pods, filter for NodeName, list nodes, choose a node fit, bind pod to Node. Start with easy logic.
5. cmd/mk-kubelet: register node object, watch pods assigned, start/stop workloads, update pod and node status. No need to
   run OCI containers in the start, use exec.Command
6. cmd/mk-controllermanager: deployment or node controller as a start. Compare desired and actual, then converge.

7. pkg/api/v1: the public api, versioned for extension
8. pkg/apimachinery: metadata, lists, watch events, status objects, conditions, api errors
9. pkg/store: the backend implementation for api server for storage

10. pkg/apiserver:
    a. server.go: boot HTTP server
    b. router.go: register routes
    c. handlers_xx.go: resource specific logic
    d. codec.go: json encode/decode
    e. admission.go: defaulting + base validation
11. pkg/client: for controllers and agents to interact with API Server (createPod, ListPod....)

## Leaving out

1. RBAC
2. API Aggregation
3. Admission Webhooks
4. CRDs
5. Informer Caches
6. Leader Election
7. CNI/CSI
8. kube-proxy
9. Services and DNS
