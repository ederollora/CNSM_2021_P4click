package types

const (
	TOOL       = iota
	DEPLOYMENT
	MACHINE
	CONTROLLER
	SWITCH
)

const (
	CERT_PASSWORD = 1
	CERT_PUBLIC_KEY_FILE = 2
	DEFAULT_TIMEOUT = 3 // second
)

const (
	ONOSPORTREST    = 8181
	ONOSOPENFLOW    = 6653
	ONOSKARAFSSH    = 8101
	ONOSPORTDEBUG   = 5005
	ONOSPORTCLUSTER = 9876
	ONOSOVSDB       = 6640
	ONOSPORTNETCONF = 830
)

const (
	PRINT_IN_TERMINAL = true
	DO_NOT_PRINT_IN_TERMINAL = false
	P4CLICK = "p4click"
	DEPENDENCIES = "dependencies"

	HEADER_DEP_FILE = "header.dep.yaml"
	CONSTANT_DEP_FILE = "constant.dep.yaml"
	DEFINE_DEP_FILE = "define.dep.yaml"
	TYPEDEF_DEP_FILE = "typedef.dep.yaml"
	METADATA_DEP_FILE = "metadata.dep.yaml"
	REPOSITORY_DIRECTORY = "/home/eoza/go/src/repository"
	HTTPS_PROTO = "https://"
	P4Z = ".p4z"
	PIPELINE = "pipeline"
	SPACE = " "
	NEWLINE = "\n"
)


type ContextType struct{
	Name string
	Ctype int
}

const (

	P4                = "p4"

	PACKET            = "packet"
	HDR               = "hdr"
	MODEL_CONFIG_FILE = "%s.config.yaml"

	CODE              = "code"
	CONFIG            = "config"
	YAML              = "yaml"


	HASH              = "#"
	OPEN_BRACKET      = "{"
	CLOSING_BRACKET   = "}"
	OPEN_PARENTHESIS  = "("
	CLOSE_PARENTHESIS = ")"
	END_STATEMENT     = ";"
	SPACES            = "%s"
	COMMA             = ","
	DOT               = "."
	BREAKLINE         = "\n"
	EQUAL             = "="
	DOUBLE_QUOTE      = "\""
	LOWER_THAN        = "<"
	BIGGER_THAN       = ">"


	TYPE_DEC       = "_t"
	START             = "start"
	ACCEPT            = "accept"

	DIRECT            = "direct"
	CONDITIONAL       = "conditional"

	HEADER_FILE       = "headers.p4"
	PARSER_FILE       = "parser.p4"
	DEPARSER_FILE     = "deparser.p4"

	MAIN              = "main"

	PARAMETER         = "%s"

	ARGUMENT_2        = "%s %s"
	ARGUMENT_3        = "%s %s %s"

	HEADER            = "header"
	HEADERS           = "headers"
	CONSTANT          = "const"
	DEFINE            = "define"
	TYPEDEF           = "typedef"
	INCLUDE           = "include"
	MODULE_CODE       = "modulecode"
	STRUCT            = "struct"
	METADATA          = "metadata"
	LOCAL_METADATA    = "local_metadata"


	BIT_TYPE          = "bit<%s>"
	TYPEDEF_TYPE      = "%s"
	ANNOTATION_LINE   = "@%s(\"%s\")"
	HEADER_LINE       = "header %s %s"
	HEADER_FIELD_LINE = "%s%s%s"+ END_STATEMENT
	STRUCT_HEADERS    = STRUCT + SPACE + PARAMETER + SPACE + OPEN_BRACKET
	ONE_HEADER_STMT   = PARAMETER + PARAMETER + PARAMETER + END_STATEMENT
	STRUCT_META       = STRUCT + SPACE + PARAMETER + SPACE + OPEN_BRACKET
	META_LINE         = PARAMETER + SPACE + PARAMETER + END_STATEMENT
	ONE_META_STMT     = PARAMETER + PARAMETER + PARAMETER + END_STATEMENT

	DEPARSER          = "deparser"
	PARSER            = "parser"
	INGRESS_MAU       = "ingressMau"
	EGRESS_MAU        = "egressMau"
	COMPUTE_CHK       = "computeChk"
	VERIFY_CHK        = "verifyChk"


	BLOCK_FUNC_HEADER = "%s %s (%s) " + OPEN_BRACKET
	STATE             = "state"
	PARSER_STATE      = "%s %s " + OPEN_BRACKET
	EXTRACT_STATEMENT = "%s.extract(%s.%s)" + END_STATEMENT
	TRANSITION        = "transition"
	TRANSITION_FUNC   = TRANSITION + SPACE + "select(%s) " + OPEN_BRACKET
	TRANSITION_COND   = "%s.%s.%s"
	TRANSITION_META   = "%s.%s"
	ONE_SELECT_TR     = "%s: %s" + END_STATEMENT
	SIMPLE_TRANSITION = TRANSITION + SPACE + "%s" + END_STATEMENT
	DEFAULT           = "default"

	CONTROL_DEC       = PARAMETER + OPEN_PARENTHESIS + CLOSE_PARENTHESIS + SPACE + PARAMETER + END_STATEMENT
	CONTROL_APPLY     = PARAMETER + DOT + APPLY + OPEN_PARENTHESIS + PARAMETER + CLOSE_PARENTHESIS + END_STATEMENT

	TYPEDEF_STATEMENT = TYPEDEF + SPACE + BIT_TYPE + SPACE + PARAMETER + END_STATEMENT

	CONST_BIT_STMT    = CONSTANT + SPACE + BIT_TYPE + SPACE + PARAMETER + SPACE + EQUAL + SPACE + PARAMETER
	CONST_TP_STMT     = CONSTANT + SPACE + PARAMETER + SPACE + PARAMETER + SPACE + EQUAL + SPACE + PARAMETER

	DEFINE_STATEMENT  = HASH + DEFINE + SPACE + PARAMETER + SPACE + PARAMETER

	DEF_INCLUDE       =  HASH + INCLUDE + SPACE + LOWER_THAN + PARAMETER + BIGGER_THAN
	INCLUDE_STATEMENT = HASH + INCLUDE + SPACE + DOUBLE_QUOTE + PARAMETER + DOUBLE_QUOTE

	SWITCH_HEADER     = PARAMETER + OPEN_PARENTHESIS
	PARAM_SWITCH      = PARAMETER + OPEN_PARENTHESIS + CLOSE_PARENTHESIS
	CLOSING_SWITCH    = CLOSE_PARENTHESIS + SPACE + PARAMETER + END_STATEMENT


	APPLY             = "apply"
	APPLY_OPEN        = APPLY + SPACE + OPEN_BRACKET

	EMIT              = "emit"
	EMIT_STATEMENT    = PACKET + DOT + EMIT + OPEN_PARENTHESIS + HDR + DOT + PARAMETER + CLOSE_PARENTHESIS + END_STATEMENT

	OK                = "OK"
	FAILED            = "FAILED"


)

