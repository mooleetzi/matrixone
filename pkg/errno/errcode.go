// Copyright 2021 Matrix Origin
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package errno

// PGSQL error code
const (
	SuccessFulCompletion                    = "00000"
	Waring                                  = "01000"
	NoData                                  = "02000"
	SQLStatementNotYetComplete              = "03000"
	ConnectionException                     = "08000"
	TriggeredActionException                = "09000"
	FeatureNotSupported                     = "0A000"
	InvalidTransactionInitiation            = "0B000"
	LocatorException                        = "0F000"
	InvalidGrantor                          = "0L000"
	InvalidRoleSpecification                = "0P000"
	DiagnosticsException                    = "0Z000"
	CaseNotFound                            = "20000"
	CardinalityViolation                    = "21000"
	DataException                           = "22000"
	IntegrityConstraintViolation            = "23000"
	InvalidCursorState                      = "24000"
	InvalidTransactionState                 = "25000"
	InvalidSQLStatementName                 = "26000"
	TriggeredDataChangeViolation            = "27000"
	InvalidAuthorizationSpecification       = "28000"
	DependentPrivilegeDescriptorsStillExist = "2B000"
	InvalidTransactionTermination           = "2D000"
	SQLRoutineException                     = "2F000"
	InvalidCursorName                       = "34000"
	ExternalRoutineException                = "38000"
	ExternalRoutineInvocationException      = "39000"
	SavepointException                      = "3B000"
	InvalidCatalogName                      = "3D000"
	InvalidSchemaName                       = "3F000"
	TransactionRollback                     = "40000"
	SyntaxErrororAccessRuleViolation        = "42000"
	SyntaxError                             = "42601"
	InsufficientPrivilege                   = "42501"
	CannotCoerce                            = "42846"
	GroupingError                           = "42803"
	WindowingError                          = "42P20"
	InvalidRecursion                        = "42P19"
	InvalidForeignKey                       = "42830"
	InvalidName                             = "42602"
	InvalidOptionValue                      = "42603"
	NameTooLong                             = "42622"
	ReservedName                            = "42939"
	DatatypeMismatch                        = "42804"
	IndeterminateDatatype                   = "42P18"
	CollationMismatch                       = "42P21"
	IndeterminateCollation                  = "42P22"
	WrongObjectType                         = "42809"
	GeneratedAlways                         = "428C9"
	UndefinedColumn                         = "42703"
	UndefinedFunction                       = "42883"
	UndefinedTable                          = "42P01"
	UndefinedParameter                      = "42P02"
	UndefinedObject                         = "42704"
	DuplicateColumn                         = "42701"
	DuplicateCursor                         = "42P03"
	DuplicateDatabase                       = "42P04"
	DuplicateFunction                       = "42723"
	DuplicatePreparedStatement              = "42P05"
	DuplicateSchema                         = "42P06"
	DuplicateTable                          = "42P07"
	DuplicateAlias                          = "42712"
	DuplicateObject                         = "42710"
	AmbiguousColumn                         = "42702"
	AmbiguousFunction                       = "42725"
	AmbiguousParameter                      = "42P08"
	AmbiguousAlias                          = "42P09"
	InvalidColumnReference                  = "42P10"
	InvalidColumnDefinition                 = "42611"
	InvalidCursorDefinition                 = "42P11"
	InvalidDatabaseDefinition               = "42P12"
	InvalidFunctionDefinition               = "42P13"
	InvalidPreparedStatementDefinition      = "42P14"
	InvalidSchemaDefinition                 = "42P15"
	InvalidTableDefinition                  = "42P16"
	InvalidObjectDefinition                 = "42P17"
	WITHCHECKOPTIONViolation                = "44000"
	InsufficientResources                   = "53000"
	ProgramLimitExceeded                    = "54000"
	ObjectNotInPrerequisiteState            = "55000"
	OperatorIntervention                    = "56000"
	SystemError                             = "58000"
	InternalError                           = "XX000"
	InvalidJsonText                         = "the JSON text is not valid"
	EmptyJsonText                           = "the JSON text is empty"
	InvalidJsonNumber                       = "the JSON number is not valid"
	InvalidJsonKeyTooLong                   = "the JSON key is too long"
	UnSupportedJsonType                     = "the JSON data type is not supported"
	InvalidJsonPath                         = "the JSON path is not valid"
	InvalidUnnestMode                       = "the unnest mode is not valid, only support 'array', 'object' and 'both'"
)
