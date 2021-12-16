package mongoerror

import "go.mongodb.org/mongo-driver/bson"

type ErrorCode int

const (
	OK                                        ErrorCode = 0
	InternalError                             ErrorCode = 1
	BadValue                                  ErrorCode = 2
	OBSOLETE_DuplicateKey                     ErrorCode = 3
	NoSuchKey                                 ErrorCode = 4
	GraphContainsCycle                        ErrorCode = 5
	HostUnreachable                           ErrorCode = 6
	HostNotFound                              ErrorCode = 7
	UnknownError                              ErrorCode = 8
	FailedToParse                             ErrorCode = 9
	CannotMutateObject                        ErrorCode = 10
	UserNotFound                              ErrorCode = 11
	UnsupportedFormat                         ErrorCode = 12
	Unauthorized                              ErrorCode = 13
	TypeMismatch                              ErrorCode = 14
	Overflow                                  ErrorCode = 15
	InvalidLength                             ErrorCode = 16
	ProtocolError                             ErrorCode = 17
	AuthenticationFailed                      ErrorCode = 18
	CannotReuseObject                         ErrorCode = 19
	IllegalOperation                          ErrorCode = 20
	EmptyArrayOperation                       ErrorCode = 21
	InvalidBSON                               ErrorCode = 22
	AlreadyInitialized                        ErrorCode = 23
	LockTimeout                               ErrorCode = 24
	RemoteValidationError                     ErrorCode = 25
	NamespaceNotFound                         ErrorCode = 26
	IndexNotFound                             ErrorCode = 27
	PathNotViable                             ErrorCode = 28
	NonExistentPath                           ErrorCode = 29
	InvalidPath                               ErrorCode = 30
	RoleNotFound                              ErrorCode = 31
	RolesNotRelated                           ErrorCode = 32
	PrivilegeNotFound                         ErrorCode = 33
	CannotBackfillArray                       ErrorCode = 34
	UserModificationFailed                    ErrorCode = 35
	RemoteChangeDetected                      ErrorCode = 36
	FileRenameFailed                          ErrorCode = 37
	FileNotOpen                               ErrorCode = 38
	FileStreamFailed                          ErrorCode = 39
	ConflictingUpdateOperators                ErrorCode = 40
	FileAlreadyOpen                           ErrorCode = 41
	LogWriteFailed                            ErrorCode = 42
	CursorNotFound                            ErrorCode = 43
	UserDataInconsistent                      ErrorCode = 45
	LockBusy                                  ErrorCode = 46
	NoMatchingDocument                        ErrorCode = 47
	NamespaceExists                           ErrorCode = 48
	InvalidRoleModification                   ErrorCode = 49
	MaxTimeMSExpired                          ErrorCode = 50
	ManualInterventionRequired                ErrorCode = 51
	DollarPrefixedFieldName                   ErrorCode = 52
	InvalidIdField                            ErrorCode = 53
	NotSingleValueField                       ErrorCode = 54
	InvalidDBRef                              ErrorCode = 55
	EmptyFieldName                            ErrorCode = 56
	DottedFieldName                           ErrorCode = 57
	RoleModificationFailed                    ErrorCode = 58
	CommandNotFound                           ErrorCode = 59
	OBSOLETE_DatabaseNotFound                 ErrorCode = 60
	ShardKeyNotFound                          ErrorCode = 61
	OplogOperationUnsupported                 ErrorCode = 62
	StaleShardVersion                         ErrorCode = 63
	WriteConcernFailed                        ErrorCode = 64
	MultipleErrorsOccurred                    ErrorCode = 65
	ImmutableField                            ErrorCode = 66
	CannotCreateIndex                         ErrorCode = 67
	IndexAlreadyExists                        ErrorCode = 68
	AuthSchemaIncompatible                    ErrorCode = 69
	ShardNotFound                             ErrorCode = 70
	ReplicaSetNotFound                        ErrorCode = 71
	InvalidOptions                            ErrorCode = 72
	InvalidNamespace                          ErrorCode = 73
	NodeNotFound                              ErrorCode = 74
	WriteConcernLegacyOK                      ErrorCode = 75
	NoReplicationEnabled                      ErrorCode = 76
	OperationIncomplete                       ErrorCode = 77
	CommandResultSchemaViolation              ErrorCode = 78
	UnknownReplWriteConcern                   ErrorCode = 79
	RoleDataInconsistent                      ErrorCode = 80
	NoMatchParseContext                       ErrorCode = 81
	NoProgressMade                            ErrorCode = 82
	RemoteResultsUnavailable                  ErrorCode = 83
	DuplicateKeyValue                         ErrorCode = 84
	IndexOptionsConflict                      ErrorCode = 85
	IndexKeySpecsConflict                     ErrorCode = 86
	CannotSplit                               ErrorCode = 87
	SplitFailed_OBSOLETE                      ErrorCode = 88
	NetworkTimeout                            ErrorCode = 89
	CallbackCanceled                          ErrorCode = 90
	ShutdownInProgress                        ErrorCode = 91
	SecondaryAheadOfPrimary                   ErrorCode = 92
	InvalidReplicaSetConfig                   ErrorCode = 93
	NotYetInitialized                         ErrorCode = 94
	NotSecondary                              ErrorCode = 95
	OperationFailed                           ErrorCode = 96
	NoProjectionFound                         ErrorCode = 97
	DBPathInUse                               ErrorCode = 98
	CannotSatisfyWriteConcern                 ErrorCode = 100
	OutdatedClient                            ErrorCode = 101
	IncompatibleAuditMetadata                 ErrorCode = 102
	NewReplicaSetConfigurationIncompatible    ErrorCode = 103
	NodeNotElectable                          ErrorCode = 104
	IncompatibleShardingMetadata              ErrorCode = 105
	DistributedClockSkewed                    ErrorCode = 106
	LockFailed                                ErrorCode = 107
	InconsistentReplicaSetNames               ErrorCode = 108
	ConfigurationInProgress                   ErrorCode = 109
	CannotInitializeNodeWithData              ErrorCode = 110
	NotExactValueField                        ErrorCode = 111
	WriteConflict                             ErrorCode = 112
	InitialSyncFailure                        ErrorCode = 113
	InitialSyncOplogSourceMissing             ErrorCode = 114
	CommandNotSupported                       ErrorCode = 115
	DocTooLargeForCapped                      ErrorCode = 116
	ConflictingOperationInProgress            ErrorCode = 117
	NamespaceNotSharded                       ErrorCode = 118
	InvalidSyncSource                         ErrorCode = 119
	OplogStartMissing                         ErrorCode = 120
	DocumentValidationFailure                 ErrorCode = 121
	OBSOLETE_ReadAfterOptimeTimeout           ErrorCode = 122
	NotAReplicaSet                            ErrorCode = 123
	IncompatibleElectionProtocol              ErrorCode = 124
	CommandFailed                             ErrorCode = 125
	RPCProtocolNegotiationFailed              ErrorCode = 126
	UnrecoverableRollbackError                ErrorCode = 127
	LockNotFound                              ErrorCode = 128
	LockStateChangeFailed                     ErrorCode = 129
	SymbolNotFound                            ErrorCode = 130
	RLPInitializationFailed                   ErrorCode = 131
	OBSOLETE_ConfigServersInconsistent        ErrorCode = 132
	FailedToSatisfyReadPreference             ErrorCode = 133
	ReadConcernMajorityNotAvailableYet        ErrorCode = 134
	StaleTerm                                 ErrorCode = 135
	CappedPositionLost                        ErrorCode = 136
	IncompatibleShardingConfigVersion         ErrorCode = 137
	RemoteOplogStale                          ErrorCode = 138
	JSInterpreterFailure                      ErrorCode = 139
	InvalidSSLConfiguration                   ErrorCode = 140
	SSLHandshakeFailed                        ErrorCode = 141
	JSUncatchableError                        ErrorCode = 142
	CursorInUse                               ErrorCode = 143
	IncompatibleCatalogManager                ErrorCode = 144
	PooledConnectionsDropped                  ErrorCode = 145
	ExceededMemoryLimit                       ErrorCode = 146
	ZLibError                                 ErrorCode = 147
	ReadConcernMajorityNotEnabled             ErrorCode = 148
	NoConfigMaster                            ErrorCode = 149
	StaleEpoch                                ErrorCode = 150
	OperationCannotBeBatched                  ErrorCode = 151
	OplogOutOfOrder                           ErrorCode = 152
	ChunkTooBig                               ErrorCode = 153
	InconsistentShardIdentity                 ErrorCode = 154
	CannotApplyOplogWhilePrimary              ErrorCode = 155
	NeedsDocumentMove                         ErrorCode = 156
	CanRepairToDowngrade                      ErrorCode = 157
	MustUpgrade                               ErrorCode = 158
	DurationOverflow                          ErrorCode = 159
	MaxStalenessOutOfRange                    ErrorCode = 160
	IncompatibleCollationVersion              ErrorCode = 161
	CollectionIsEmpty                         ErrorCode = 162
	ZoneStillInUse                            ErrorCode = 163
	InitialSyncActive                         ErrorCode = 164
	ViewDepthLimitExceeded                    ErrorCode = 165
	CommandNotSupportedOnView                 ErrorCode = 166
	OptionNotSupportedOnView                  ErrorCode = 167
	InvalidPipelineOperator                   ErrorCode = 168
	CommandOnShardedViewNotSupportedOnMongod  ErrorCode = 169
	TooManyMatchingDocuments                  ErrorCode = 170
	CannotIndexParallelArrays                 ErrorCode = 171
	TransportSessionClosed                    ErrorCode = 172
	TransportSessionNotFound                  ErrorCode = 173
	TransportSessionUnknown                   ErrorCode = 174
	QueryPlanKilled                           ErrorCode = 175
	FileOpenFailed                            ErrorCode = 176
	ZoneNotFound                              ErrorCode = 177
	RangeOverlapConflict                      ErrorCode = 178
	WindowsPdhError                           ErrorCode = 179
	BadPerfCounterPath                        ErrorCode = 180
	AmbiguousIndexKeyPattern                  ErrorCode = 181
	InvalidViewDefinition                     ErrorCode = 182
	ClientMetadataMissingField                ErrorCode = 183
	ClientMetadataAppNameTooLarge             ErrorCode = 184
	ClientMetadataDocumentTooLarge            ErrorCode = 185
	ClientMetadataCannotBeMutated             ErrorCode = 186
	LinearizableReadConcernError              ErrorCode = 187
	IncompatibleServerVersion                 ErrorCode = 188
	PrimarySteppedDown                        ErrorCode = 189
	MasterSlaveConnectionFailure              ErrorCode = 190
	OBSOLETE_BalancerLostDistributedLock      ErrorCode = 191
	FailPointEnabled                          ErrorCode = 192
	NoShardingEnabled                         ErrorCode = 193
	BalancerInterrupted                       ErrorCode = 194
	ViewPipelineMaxSizeExceeded               ErrorCode = 195
	InvalidIndexSpecificationOption           ErrorCode = 197
	OBSOLETE_ReceivedOpReplyMessage           ErrorCode = 198
	ReplicaSetMonitorRemoved                  ErrorCode = 199
	ChunkRangeCleanupPending                  ErrorCode = 200
	CannotBuildIndexKeys                      ErrorCode = 201
	NetworkInterfaceExceededTimeLimit         ErrorCode = 202
	ShardingStateNotInitialized               ErrorCode = 203
	TimeProofMismatch                         ErrorCode = 204
	ClusterTimeFailsRateLimiter               ErrorCode = 205
	NoSuchSession                             ErrorCode = 206
	InvalidUUID                               ErrorCode = 207
	TooManyLocks                              ErrorCode = 208
	StaleClusterTime                          ErrorCode = 209
	CannotVerifyAndSignLogicalTime            ErrorCode = 210
	KeyNotFound                               ErrorCode = 211
	IncompatibleRollbackAlgorithm             ErrorCode = 212
	DuplicateSession                          ErrorCode = 213
	AuthenticationRestrictionUnmet            ErrorCode = 214
	DatabaseDropPending                       ErrorCode = 215
	ElectionInProgress                        ErrorCode = 216
	IncompleteTransactionHistory              ErrorCode = 217
	UpdateOperationFailed                     ErrorCode = 218
	FTDCPathNotSet                            ErrorCode = 219
	FTDCPathAlreadySet                        ErrorCode = 220
	IndexModified                             ErrorCode = 221
	CloseChangeStream                         ErrorCode = 222
	IllegalOpMsgFlag                          ErrorCode = 223
	QueryFeatureNotAllowed                    ErrorCode = 224
	TransactionTooOld                         ErrorCode = 225
	AtomicityFailure                          ErrorCode = 226
	CannotImplicitlyCreateCollection          ErrorCode = 227
	SessionTransferIncomplete                 ErrorCode = 228
	MustDowngrade                             ErrorCode = 229
	DNSHostNotFound                           ErrorCode = 230
	DNSProtocolError                          ErrorCode = 231
	MaxSubPipelineDepthExceeded               ErrorCode = 232
	TooManyDocumentSequences                  ErrorCode = 233
	RetryChangeStream                         ErrorCode = 234
	InternalErrorNotSupported                 ErrorCode = 235
	ForTestingErrorExtraInfo                  ErrorCode = 236
	CursorKilled                              ErrorCode = 237
	NotImplemented                            ErrorCode = 238
	SnapshotTooOld                            ErrorCode = 239
	DNSRecordTypeMismatch                     ErrorCode = 240
	ConversionFailure                         ErrorCode = 241
	CannotCreateCollection                    ErrorCode = 242
	IncompatibleWithUpgradedServer            ErrorCode = 243
	NOT_YET_AVAILABLE_TransactionAborted      ErrorCode = 244
	BrokenPromise                             ErrorCode = 245
	SnapshotUnavailable                       ErrorCode = 246
	ProducerConsumerQueueBatchTooLarge        ErrorCode = 247
	ProducerConsumerQueueEndClosed            ErrorCode = 248
	StaleDbVersion                            ErrorCode = 249
	StaleChunkHistory                         ErrorCode = 250
	NoSuchTransaction                         ErrorCode = 251
	ReentrancyNotAllowed                      ErrorCode = 252
	FreeMonHttpInFlight                       ErrorCode = 253
	FreeMonHttpTemporaryFailure               ErrorCode = 254
	FreeMonHttpPermanentFailure               ErrorCode = 255
	TransactionCommitted                      ErrorCode = 256
	TransactionTooLarge                       ErrorCode = 257
	UnknownFeatureCompatibilityVersion        ErrorCode = 258
	KeyedExecutorRetry                        ErrorCode = 259
	InvalidResumeToken                        ErrorCode = 260
	TooManyLogicalSessions                    ErrorCode = 261
	ExceededTimeLimit                         ErrorCode = 262
	OperationNotSupportedInTransaction        ErrorCode = 263
	TooManyFilesOpen                          ErrorCode = 264
	FailPointSetFailed                        ErrorCode = 266
	DataModifiedByRepair                      ErrorCode = 269
	RepairedReplicaSetNode                    ErrorCode = 270
	SocketException                           ErrorCode = 9001
	OBSOLETE_RecvStaleConfig                  ErrorCode = 9996
	CannotGrowDocumentInCappedNamespace       ErrorCode = 10003
	NotMaster                                 ErrorCode = 10107
	BSONObjectTooLarge                        ErrorCode = 10334
	DuplicateKey                              ErrorCode = 11000
	InterruptedAtShutdown                     ErrorCode = 11600
	Interrupted                               ErrorCode = 11601
	InterruptedDueToReplStateChange           ErrorCode = 11602
	BackgroundOperationInProgressForDatabase  ErrorCode = 12586
	BackgroundOperationInProgressForNamespace ErrorCode = 12587
	OBSOLETE_PrepareConfigsFailed             ErrorCode = 13104
	DatabaseDifferCase                        ErrorCode = 13297
	ShardKeyTooBig                            ErrorCode = 13334
	StaleConfig                               ErrorCode = 13388
	NotMasterNoSlaveOk                        ErrorCode = 13435
	NotMasterOrSecondary                      ErrorCode = 13436
	OutOfDiskSpace                            ErrorCode = 14031
	KeyTooLong                                ErrorCode = 17280
)

func (c ErrorCode) String() string {
	switch c {
	case 0:
		return "OK"
	case 1:
		return "InternalError"
	case 2:
		return "BadValue"
	case 3:
		return "OBSOLETE_DuplicateKey"
	case 4:
		return "NoSuchKey"
	case 5:
		return "GraphContainsCycle"
	case 6:
		return "HostUnreachable"
	case 7:
		return "HostNotFound"
	case 8:
		return "UnknownError"
	case 9:
		return "FailedToParse"
	case 10:
		return "CannotMutateObject"
	case 11:
		return "UserNotFound"
	case 12:
		return "UnsupportedFormat"
	case 13:
		return "Unauthorized"
	case 14:
		return "TypeMismatch"
	case 15:
		return "Overflow"
	case 16:
		return "InvalidLength"
	case 17:
		return "ProtocolError"
	case 18:
		return "AuthenticationFailed"
	case 19:
		return "CannotReuseObject"
	case 20:
		return "IllegalOperation"
	case 21:
		return "EmptyArrayOperation"
	case 22:
		return "InvalidBSON"
	case 23:
		return "AlreadyInitialized"
	case 24:
		return "LockTimeout"
	case 25:
		return "RemoteValidationError"
	case 26:
		return "NamespaceNotFound"
	case 27:
		return "IndexNotFound"
	case 28:
		return "PathNotViable"
	case 29:
		return "NonExistentPath"
	case 30:
		return "InvalidPath"
	case 31:
		return "RoleNotFound"
	case 32:
		return "RolesNotRelated"
	case 33:
		return "PrivilegeNotFound"
	case 34:
		return "CannotBackfillArray"
	case 35:
		return "UserModificationFailed"
	case 36:
		return "RemoteChangeDetected"
	case 37:
		return "FileRenameFailed"
	case 38:
		return "FileNotOpen"
	case 39:
		return "FileStreamFailed"
	case 40:
		return "ConflictingUpdateOperators"
	case 41:
		return "FileAlreadyOpen"
	case 42:
		return "LogWriteFailed"
	case 43:
		return "CursorNotFound"
	case 45:
		return "UserDataInconsistent"
	case 46:
		return "LockBusy"
	case 47:
		return "NoMatchingDocument"
	case 48:
		return "NamespaceExists"
	case 49:
		return "InvalidRoleModification"
	case 50:
		return "MaxTimeMSExpired"
	case 51:
		return "ManualInterventionRequired"
	case 52:
		return "DollarPrefixedFieldName"
	case 53:
		return "InvalidIdField"
	case 54:
		return "NotSingleValueField"
	case 55:
		return "InvalidDBRef"
	case 56:
		return "EmptyFieldName"
	case 57:
		return "DottedFieldName"
	case 58:
		return "RoleModificationFailed"
	case 59:
		return "CommandNotFound"
	case 60:
		return "OBSOLETE_DatabaseNotFound"
	case 61:
		return "ShardKeyNotFound"
	case 62:
		return "OplogOperationUnsupported"
	case 63:
		return "StaleShardVersion"
	case 64:
		return "WriteConcernFailed"
	case 65:
		return "MultipleErrorsOccurred"
	case 66:
		return "ImmutableField"
	case 67:
		return "CannotCreateIndex"
	case 68:
		return "IndexAlreadyExists"
	case 69:
		return "AuthSchemaIncompatible"
	case 70:
		return "ShardNotFound"
	case 71:
		return "ReplicaSetNotFound"
	case 72:
		return "InvalidOptions"
	case 73:
		return "InvalidNamespace"
	case 74:
		return "NodeNotFound"
	case 75:
		return "WriteConcernLegacyOK"
	case 76:
		return "NoReplicationEnabled"
	case 77:
		return "OperationIncomplete"
	case 78:
		return "CommandResultSchemaViolation"
	case 79:
		return "UnknownReplWriteConcern"
	case 80:
		return "RoleDataInconsistent"
	case 81:
		return "NoMatchParseContext"
	case 82:
		return "NoProgressMade"
	case 83:
		return "RemoteResultsUnavailable"
	case 84:
		return "DuplicateKeyValue"
	case 85:
		return "IndexOptionsConflict"
	case 86:
		return "IndexKeySpecsConflict"
	case 87:
		return "CannotSplit"
	case 88:
		return "SplitFailed_OBSOLETE"
	case 89:
		return "NetworkTimeout"
	case 90:
		return "CallbackCanceled"
	case 91:
		return "ShutdownInProgress"
	case 92:
		return "SecondaryAheadOfPrimary"
	case 93:
		return "InvalidReplicaSetConfig"
	case 94:
		return "NotYetInitialized"
	case 95:
		return "NotSecondary"
	case 96:
		return "OperationFailed"
	case 97:
		return "NoProjectionFound"
	case 98:
		return "DBPathInUse"
	case 100:
		return "CannotSatisfyWriteConcern"
	case 101:
		return "OutdatedClient"
	case 102:
		return "IncompatibleAuditMetadata"
	case 103:
		return "NewReplicaSetConfigurationIncompatible"
	case 104:
		return "NodeNotElectable"
	case 105:
		return "IncompatibleShardingMetadata"
	case 106:
		return "DistributedClockSkewed"
	case 107:
		return "LockFailed"
	case 108:
		return "InconsistentReplicaSetNames"
	case 109:
		return "ConfigurationInProgress"
	case 110:
		return "CannotInitializeNodeWithData"
	case 111:
		return "NotExactValueField"
	case 112:
		return "WriteConflict"
	case 113:
		return "InitialSyncFailure"
	case 114:
		return "InitialSyncOplogSourceMissing"
	case 115:
		return "CommandNotSupported"
	case 116:
		return "DocTooLargeForCapped"
	case 117:
		return "ConflictingOperationInProgress"
	case 118:
		return "NamespaceNotSharded"
	case 119:
		return "InvalidSyncSource"
	case 120:
		return "OplogStartMissing"
	case 121:
		return "DocumentValidationFailure"
	case 122:
		return "OBSOLETE_ReadAfterOptimeTimeout"
	case 123:
		return "NotAReplicaSet"
	case 124:
		return "IncompatibleElectionProtocol"
	case 125:
		return "CommandFailed"
	case 126:
		return "RPCProtocolNegotiationFailed"
	case 127:
		return "UnrecoverableRollbackError"
	case 128:
		return "LockNotFound"
	case 129:
		return "LockStateChangeFailed"
	case 130:
		return "SymbolNotFound"
	case 131:
		return "RLPInitializationFailed"
	case 132:
		return "OBSOLETE_ConfigServersInconsistent"
	case 133:
		return "FailedToSatisfyReadPreference"
	case 134:
		return "ReadConcernMajorityNotAvailableYet"
	case 135:
		return "StaleTerm"
	case 136:
		return "CappedPositionLost"
	case 137:
		return "IncompatibleShardingConfigVersion"
	case 138:
		return "RemoteOplogStale"
	case 139:
		return "JSInterpreterFailure"
	case 140:
		return "InvalidSSLConfiguration"
	case 141:
		return "SSLHandshakeFailed"
	case 142:
		return "JSUncatchableError"
	case 143:
		return "CursorInUse"
	case 144:
		return "IncompatibleCatalogManager"
	case 145:
		return "PooledConnectionsDropped"
	case 146:
		return "ExceededMemoryLimit"
	case 147:
		return "ZLibError"
	case 148:
		return "ReadConcernMajorityNotEnabled"
	case 149:
		return "NoConfigMaster"
	case 150:
		return "StaleEpoch"
	case 151:
		return "OperationCannotBeBatched"
	case 152:
		return "OplogOutOfOrder"
	case 153:
		return "ChunkTooBig"
	case 154:
		return "InconsistentShardIdentity"
	case 155:
		return "CannotApplyOplogWhilePrimary"
	case 156:
		return "NeedsDocumentMove"
	case 157:
		return "CanRepairToDowngrade"
	case 158:
		return "MustUpgrade"
	case 159:
		return "DurationOverflow"
	case 160:
		return "MaxStalenessOutOfRange"
	case 161:
		return "IncompatibleCollationVersion"
	case 162:
		return "CollectionIsEmpty"
	case 163:
		return "ZoneStillInUse"
	case 164:
		return "InitialSyncActive"
	case 165:
		return "ViewDepthLimitExceeded"
	case 166:
		return "CommandNotSupportedOnView"
	case 167:
		return "OptionNotSupportedOnView"
	case 168:
		return "InvalidPipelineOperator"
	case 169:
		return "CommandOnShardedViewNotSupportedOnMongod"
	case 170:
		return "TooManyMatchingDocuments"
	case 171:
		return "CannotIndexParallelArrays"
	case 172:
		return "TransportSessionClosed"
	case 173:
		return "TransportSessionNotFound"
	case 174:
		return "TransportSessionUnknown"
	case 175:
		return "QueryPlanKilled"
	case 176:
		return "FileOpenFailed"
	case 177:
		return "ZoneNotFound"
	case 178:
		return "RangeOverlapConflict"
	case 179:
		return "WindowsPdhError"
	case 180:
		return "BadPerfCounterPath"
	case 181:
		return "AmbiguousIndexKeyPattern"
	case 182:
		return "InvalidViewDefinition"
	case 183:
		return "ClientMetadataMissingField"
	case 184:
		return "ClientMetadataAppNameTooLarge"
	case 185:
		return "ClientMetadataDocumentTooLarge"
	case 186:
		return "ClientMetadataCannotBeMutated"
	case 187:
		return "LinearizableReadConcernError"
	case 188:
		return "IncompatibleServerVersion"
	case 189:
		return "PrimarySteppedDown"
	case 190:
		return "MasterSlaveConnectionFailure"
	case 191:
		return "OBSOLETE_BalancerLostDistributedLock"
	case 192:
		return "FailPointEnabled"
	case 193:
		return "NoShardingEnabled"
	case 194:
		return "BalancerInterrupted"
	case 195:
		return "ViewPipelineMaxSizeExceeded"
	case 197:
		return "InvalidIndexSpecificationOption"
	case 198:
		return "OBSOLETE_ReceivedOpReplyMessage"
	case 199:
		return "ReplicaSetMonitorRemoved"
	case 200:
		return "ChunkRangeCleanupPending"
	case 201:
		return "CannotBuildIndexKeys"
	case 202:
		return "NetworkInterfaceExceededTimeLimit"
	case 203:
		return "ShardingStateNotInitialized"
	case 204:
		return "TimeProofMismatch"
	case 205:
		return "ClusterTimeFailsRateLimiter"
	case 206:
		return "NoSuchSession"
	case 207:
		return "InvalidUUID"
	case 208:
		return "TooManyLocks"
	case 209:
		return "StaleClusterTime"
	case 210:
		return "CannotVerifyAndSignLogicalTime"
	case 211:
		return "KeyNotFound"
	case 212:
		return "IncompatibleRollbackAlgorithm"
	case 213:
		return "DuplicateSession"
	case 214:
		return "AuthenticationRestrictionUnmet"
	case 215:
		return "DatabaseDropPending"
	case 216:
		return "ElectionInProgress"
	case 217:
		return "IncompleteTransactionHistory"
	case 218:
		return "UpdateOperationFailed"
	case 219:
		return "FTDCPathNotSet"
	case 220:
		return "FTDCPathAlreadySet"
	case 221:
		return "IndexModified"
	case 222:
		return "CloseChangeStream"
	case 223:
		return "IllegalOpMsgFlag"
	case 224:
		return "QueryFeatureNotAllowed"
	case 225:
		return "TransactionTooOld"
	case 226:
		return "AtomicityFailure"
	case 227:
		return "CannotImplicitlyCreateCollection"
	case 228:
		return "SessionTransferIncomplete"
	case 229:
		return "MustDowngrade"
	case 230:
		return "DNSHostNotFound"
	case 231:
		return "DNSProtocolError"
	case 232:
		return "MaxSubPipelineDepthExceeded"
	case 233:
		return "TooManyDocumentSequences"
	case 234:
		return "RetryChangeStream"
	case 235:
		return "InternalErrorNotSupported"
	case 236:
		return "ForTestingErrorExtraInfo"
	case 237:
		return "CursorKilled"
	case 238:
		return "NotImplemented"
	case 239:
		return "SnapshotTooOld"
	case 240:
		return "DNSRecordTypeMismatch"
	case 241:
		return "ConversionFailure"
	case 242:
		return "CannotCreateCollection"
	case 243:
		return "IncompatibleWithUpgradedServer"
	case 244:
		return "NOT_YET_AVAILABLE_TransactionAborted"
	case 245:
		return "BrokenPromise"
	case 246:
		return "SnapshotUnavailable"
	case 247:
		return "ProducerConsumerQueueBatchTooLarge"
	case 248:
		return "ProducerConsumerQueueEndClosed"
	case 249:
		return "StaleDbVersion"
	case 250:
		return "StaleChunkHistory"
	case 251:
		return "NoSuchTransaction"
	case 252:
		return "ReentrancyNotAllowed"
	case 253:
		return "FreeMonHttpInFlight"
	case 254:
		return "FreeMonHttpTemporaryFailure"
	case 255:
		return "FreeMonHttpPermanentFailure"
	case 256:
		return "TransactionCommitted"
	case 257:
		return "TransactionTooLarge"
	case 258:
		return "UnknownFeatureCompatibilityVersion"
	case 259:
		return "KeyedExecutorRetry"
	case 260:
		return "InvalidResumeToken"
	case 261:
		return "TooManyLogicalSessions"
	case 262:
		return "ExceededTimeLimit"
	case 263:
		return "OperationNotSupportedInTransaction"
	case 264:
		return "TooManyFilesOpen"
	case 266:
		return "FailPointSetFailed"
	case 269:
		return "DataModifiedByRepair"
	case 270:
		return "RepairedReplicaSetNode"
	case 9001:
		return "SocketException"
	case 9996:
		return "OBSOLETE_RecvStaleConfig"
	case 10003:
		return "CannotGrowDocumentInCappedNamespace"
	case 10107:
		return "NotMaster"
	case 10334:
		return "BSONObjectTooLarge"
	case 11000:
		return "DuplicateKey"
	case 11600:
		return "InterruptedAtShutdown"
	case 11601:
		return "Interrupted"
	case 11602:
		return "InterruptedDueToReplStateChange"
	case 12586:
		return "BackgroundOperationInProgressForDatabase"
	case 12587:
		return "BackgroundOperationInProgressForNamespace"
	case 13104:
		return "OBSOLETE_PrepareConfigsFailed"
	case 13297:
		return "DatabaseDifferCase"
	case 13334:
		return "ShardKeyTooBig"
	case 13388:
		return "StaleConfig"
	case 13435:
		return "NotMasterNoSlaveOk"
	case 13436:
		return "NotMasterOrSecondary"
	case 14031:
		return "OutOfDiskSpace"
	case 17280:
		return "KeyTooLong"
	default:
		panic("Unknown")
	}
}

func (c ErrorCode) ErrMessage(msg string) bson.D {
	r := bson.D{{"ok", 0}, {"errmsg", msg}, {"code", int(c)}, {"codeName", c.String()}}
	return r
}

func IsNetworkError(e ErrorCode) bool {
	switch e {
	case HostUnreachable:
		return true
	case HostNotFound:
		return true
	case NetworkTimeout:
		return true
	case SocketException:
		return true
	}
	return false
}

func IsInterruption(e ErrorCode) bool {
	switch e {
	case Interrupted:
		return true
	case InterruptedAtShutdown:
		return true
	case InterruptedDueToReplStateChange:
		return true
	case ExceededTimeLimit:
		return true
	case MaxTimeMSExpired:
		return true
	case CursorKilled:
		return true
	case LockTimeout:
		return true
	}
	return false
}

func IsNotMasterError(e ErrorCode) bool {
	switch e {
	case NotMaster:
		return true
	case NotMasterNoSlaveOk:
		return true
	case NotMasterOrSecondary:
		return true
	case InterruptedDueToReplStateChange:
		return true
	case PrimarySteppedDown:
		return true
	}
	return false
}

func IsStaleShardVersionError(e ErrorCode) bool {
	switch e {
	case StaleConfig:
		return true
	case StaleShardVersion:
		return true
	case StaleEpoch:
		return true
	}
	return false
}

func IsNeedRetargettingError(e ErrorCode) bool {
	switch e {
	case StaleConfig:
		return true
	case StaleShardVersion:
		return true
	case StaleEpoch:
		return true
	case CannotImplicitlyCreateCollection:
		return true
	}
	return false
}

func IsWriteConcernError(e ErrorCode) bool {
	switch e {
	case WriteConcernFailed:
		return true
	case WriteConcernLegacyOK:
		return true
	case UnknownReplWriteConcern:
		return true
	case CannotSatisfyWriteConcern:
		return true
	}
	return false
}

func IsShutdownError(e ErrorCode) bool {
	switch e {
	case ShutdownInProgress:
		return true
	case InterruptedAtShutdown:
		return true
	}
	return false
}

func IsConnectionFatalMessageParseError(e ErrorCode) bool {
	switch e {
	case IllegalOpMsgFlag:
		return true
	case TooManyDocumentSequences:
		return true
	}
	return false
}

func IsExceededTimeLimitError(e ErrorCode) bool {
	switch e {
	case ExceededTimeLimit:
		return true
	case MaxTimeMSExpired:
		return true
	case NetworkInterfaceExceededTimeLimit:
		return true
	}
	return false
}

func IsSnapshotError(e ErrorCode) bool {
	switch e {
	case SnapshotTooOld:
		return true
	case SnapshotUnavailable:
		return true
	case StaleChunkHistory:
		return true
	}
	return false
}
