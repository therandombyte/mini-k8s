## Learning Roadmap (rewiring the Control plane ideas)

### Phase 1: Single Binary API Server + In-memory store + Pod object only

Building the API Server first as it defines the Object contract, and other components to consume the contract through watches and status updates

### Phase 2: Separate Scheduler and Node Agent (kubelet like), add watch stream, add Pos Status udpate

### Phase 3: Add Simple Deployment Controller and rolling replacement logic

### Phase 4: Add Persistence, auth, basic service abstraction

## Scaffolding

cmd: for binaries
cmd/mk-apiserver: load config, initialize store, initialize watches, server HTTP, persist objects
cmd/mkctl: a layer over api server to debug the system faster, like kubectl
cmd/mk-scheduler: watch pods, filter for NodeName, list nodes, choose a node fit, bind pod to Node. Start with easy logic
cmd/mk-kubelet: register node object, watch pods assigned, start/stop workloads, update pod and node status. No need to
run OCI containers in the start, use exec.Command
cmd/mk-controllermanager: deployment or node controller as a start. Compare desired and actual, then converge.

pkg/api/v1: the public api, versioned for extension
pkg/apimachinery: metadata, lists, watch events, status objects, conditions, api errors
pkg/store: the backend implementation for api server for storage

pkg/apiserver:
server.go: boot HTTP server
router.go: register routes
handlers_xx.go: resource specific logic
codec.go: json encode/decode
admission.go: defaulting + base validation
pkg/client: for controllers and agents to interact with API Server (createPod, ListPod....)

## Leaving out

RBAC
API Aggregation
Admission Webhooks
CRDs
Informer Caches
Leader Election
CNI/CSI
kube-proxy
Services and DNS
