# Protocol Documentation
<a name="top"></a>

## Table of Contents

- [rpc/v1/api.proto](#rpc/v1/api.proto)
    - [AddPeerReq](#.AddPeerReq)
    - [BanPeerReq](#.BanPeerReq)
    - [BlobInfoReq](#.BlobInfoReq)
    - [BlobInfoRes](#.BlobInfoRes)
    - [CheckoutReq](#.CheckoutReq)
    - [CheckoutRes](#.CheckoutRes)
    - [CommitReq](#.CommitReq)
    - [CommitRes](#.CommitRes)
    - [Empty](#.Empty)
    - [GetNamesReq](#.GetNamesReq)
    - [GetNamesRes](#.GetNamesRes)
    - [GetStatusRes](#.GetStatusRes)
    - [ListBlobInfoReq](#.ListBlobInfoReq)
    - [ListPeersReq](#.ListPeersReq)
    - [ListPeersRes](#.ListPeersRes)
    - [PreCommitReq](#.PreCommitReq)
    - [PreCommitRes](#.PreCommitRes)
    - [ReadAtReq](#.ReadAtReq)
    - [ReadAtRes](#.ReadAtRes)
    - [SendUpdateReq](#.SendUpdateReq)
    - [SendUpdateRes](#.SendUpdateRes)
    - [TruncateReq](#.TruncateReq)
    - [TruncateRes](#.TruncateRes)
    - [UnbanPeerReq](#.UnbanPeerReq)
    - [WriteAtReq](#.WriteAtReq)
    - [WriteAtRes](#.WriteAtRes)
  
    - [Footnotev1](#.Footnotev1)
  
- [Scalar Value Types](#scalar-value-types)



<a name="rpc/v1/api.proto"></a>
<p align="right"><a href="#top">Top</a></p>

## rpc/v1/api.proto



<a name=".AddPeerReq"></a>

### AddPeerReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| peerID | [bytes](#bytes) |  |  |
| ip | [string](#string) |  |  |
| verifyPeerID | [bool](#bool) |  |  |






<a name=".BanPeerReq"></a>

### BanPeerReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ip | [string](#string) |  |  |
| durationMS | [uint32](#uint32) |  |  |






<a name=".BlobInfoReq"></a>

### BlobInfoReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name=".BlobInfoRes"></a>

### BlobInfoRes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| publicKey | [bytes](#bytes) |  |  |
| importHeight | [uint32](#uint32) |  |  |
| timestamp | [uint64](#uint64) |  |  |
| merkleRoot | [bytes](#bytes) |  |  |
| reservedRoot | [bytes](#bytes) |  |  |
| receivedAt | [uint64](#uint64) |  |  |
| signature | [bytes](#bytes) |  |  |
| timebank | [uint32](#uint32) |  |  |






<a name=".CheckoutReq"></a>

### CheckoutReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name=".CheckoutRes"></a>

### CheckoutRes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| txID | [uint32](#uint32) |  |  |






<a name=".CommitReq"></a>

### CommitReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| txID | [uint32](#uint32) |  |  |
| timestamp | [uint64](#uint64) |  |  |
| signature | [bytes](#bytes) |  |  |
| broadcast | [bool](#bool) |  |  |






<a name=".CommitRes"></a>

### CommitRes







<a name=".Empty"></a>

### Empty







<a name=".GetNamesReq"></a>

### GetNamesReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| start | [string](#string) |  |  |
| count | [uint32](#uint32) |  |  |






<a name=".GetNamesRes"></a>

### GetNamesRes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| publicKey | [bytes](#bytes) |  |  |






<a name=".GetStatusRes"></a>

### GetStatusRes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| peerID | [bytes](#bytes) |  |  |
| peerCount | [uint32](#uint32) |  |  |
| headerCount | [uint32](#uint32) |  |  |
| txBytes | [uint64](#uint64) |  |  |
| rxBytes | [uint64](#uint64) |  |  |






<a name=".ListBlobInfoReq"></a>

### ListBlobInfoReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| start | [string](#string) |  |  |






<a name=".ListPeersReq"></a>

### ListPeersReq







<a name=".ListPeersRes"></a>

### ListPeersRes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| peerID | [bytes](#bytes) |  |  |
| ip | [string](#string) |  |  |
| banned | [bool](#bool) |  |  |
| connected | [bool](#bool) |  |  |
| txBytes | [uint64](#uint64) |  |  |
| rxBytes | [uint64](#uint64) |  |  |
| whitelisted | [bool](#bool) |  |  |






<a name=".PreCommitReq"></a>

### PreCommitReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| txID | [uint32](#uint32) |  |  |






<a name=".PreCommitRes"></a>

### PreCommitRes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| merkleRoot | [bytes](#bytes) |  |  |






<a name=".ReadAtReq"></a>

### ReadAtReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |
| offset | [uint32](#uint32) |  |  |
| len | [uint32](#uint32) |  |  |






<a name=".ReadAtRes"></a>

### ReadAtRes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| offset | [uint32](#uint32) |  |  |
| data | [bytes](#bytes) |  |  |






<a name=".SendUpdateReq"></a>

### SendUpdateReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| name | [string](#string) |  |  |






<a name=".SendUpdateRes"></a>

### SendUpdateRes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| recipientCount | [uint32](#uint32) |  |  |






<a name=".TruncateReq"></a>

### TruncateReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| txID | [uint32](#uint32) |  |  |






<a name=".TruncateRes"></a>

### TruncateRes







<a name=".UnbanPeerReq"></a>

### UnbanPeerReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| ip | [string](#string) |  |  |






<a name=".WriteAtReq"></a>

### WriteAtReq



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| txID | [uint32](#uint32) |  |  |
| offset | [uint32](#uint32) |  |  |
| data | [bytes](#bytes) |  |  |






<a name=".WriteAtRes"></a>

### WriteAtRes



| Field | Type | Label | Description |
| ----- | ---- | ----- | ----------- |
| bytesWritten | [uint32](#uint32) |  |  |
| writeErr | [string](#string) |  |  |





 

 

 


<a name=".Footnotev1"></a>

### Footnotev1


| Method Name | Request Type | Response Type | Description |
| ----------- | ------------ | ------------- | ------------|
| GetStatus | [.Empty](#Empty) | [.GetStatusRes](#GetStatusRes) |  |
| AddPeer | [.AddPeerReq](#AddPeerReq) | [.Empty](#Empty) |  |
| BanPeer | [.BanPeerReq](#BanPeerReq) | [.Empty](#Empty) |  |
| UnbanPeer | [.UnbanPeerReq](#UnbanPeerReq) | [.Empty](#Empty) |  |
| ListPeers | [.ListPeersReq](#ListPeersReq) | [.ListPeersRes](#ListPeersRes) stream |  |
| Checkout | [.CheckoutReq](#CheckoutReq) | [.CheckoutRes](#CheckoutRes) |  |
| WriteAt | [.WriteAtReq](#WriteAtReq) | [.WriteAtRes](#WriteAtRes) |  |
| Truncate | [.TruncateReq](#TruncateReq) | [.Empty](#Empty) |  |
| PreCommit | [.PreCommitReq](#PreCommitReq) | [.PreCommitRes](#PreCommitRes) |  |
| Commit | [.CommitReq](#CommitReq) | [.CommitRes](#CommitRes) |  |
| ReadAt | [.ReadAtReq](#ReadAtReq) | [.ReadAtRes](#ReadAtRes) |  |
| GetBlobInfo | [.BlobInfoReq](#BlobInfoReq) | [.BlobInfoRes](#BlobInfoRes) |  |
| ListBlobInfo | [.ListBlobInfoReq](#ListBlobInfoReq) | [.BlobInfoRes](#BlobInfoRes) stream |  |
| SendUpdate | [.SendUpdateReq](#SendUpdateReq) | [.SendUpdateRes](#SendUpdateRes) |  |

 



## Scalar Value Types

| .proto Type | Notes | C++ | Java | Python | Go | C# | PHP | Ruby |
| ----------- | ----- | --- | ---- | ------ | -- | -- | --- | ---- |
| <a name="double" /> double |  | double | double | float | float64 | double | float | Float |
| <a name="float" /> float |  | float | float | float | float32 | float | float | Float |
| <a name="int32" /> int32 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint32 instead. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="int64" /> int64 | Uses variable-length encoding. Inefficient for encoding negative numbers – if your field is likely to have negative values, use sint64 instead. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="uint32" /> uint32 | Uses variable-length encoding. | uint32 | int | int/long | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="uint64" /> uint64 | Uses variable-length encoding. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum or Fixnum (as required) |
| <a name="sint32" /> sint32 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int32s. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sint64" /> sint64 | Uses variable-length encoding. Signed int value. These more efficiently encode negative numbers than regular int64s. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="fixed32" /> fixed32 | Always four bytes. More efficient than uint32 if values are often greater than 2^28. | uint32 | int | int | uint32 | uint | integer | Bignum or Fixnum (as required) |
| <a name="fixed64" /> fixed64 | Always eight bytes. More efficient than uint64 if values are often greater than 2^56. | uint64 | long | int/long | uint64 | ulong | integer/string | Bignum |
| <a name="sfixed32" /> sfixed32 | Always four bytes. | int32 | int | int | int32 | int | integer | Bignum or Fixnum (as required) |
| <a name="sfixed64" /> sfixed64 | Always eight bytes. | int64 | long | int/long | int64 | long | integer/string | Bignum |
| <a name="bool" /> bool |  | bool | boolean | boolean | bool | bool | boolean | TrueClass/FalseClass |
| <a name="string" /> string | A string must always contain UTF-8 encoded or 7-bit ASCII text. | string | String | str/unicode | string | string | string | String (UTF-8) |
| <a name="bytes" /> bytes | May contain any arbitrary sequence of bytes. | string | ByteString | str | []byte | ByteString | string | String (ASCII-8BIT) |

