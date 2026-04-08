## APIMachinery

### Holds the generic pieces: Metadata, Lists, Watch Events, Status Objects, Conditions, Reusable API Errors as Kubernetes APIs rely heavily on common objet envelopes and resource metadata

meta.go will be the first file created.

- Contains the metadata for an object (split across TypeMeta and ObjectMeta)
- Metadata for a List query (invoked by controller/watch/cache)
- Condition of object
- Status of API call
