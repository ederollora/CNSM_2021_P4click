package pipeline

import (
	"archive/zip"
	"bufio"
	"fmt"
	"gopkg.in/yaml.v2"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"p4click/errors"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	. "p4click/config"
	. "p4click/types"
	. "p4click/context"
)



func BuildPipeline(modules []string, deploymentDir, target, model string){

	modelConfig := parseModelConfiguration(deploymentDir, model)

	pipelineDir := filepath.Join(deploymentDir, PIPELINE)
	os.MkdirAll(pipelineDir, os.ModePerm)

	includeDir := filepath.Join(pipelineDir, INCLUDE)
	os.MkdirAll(includeDir, os.ModePerm)

	moduleDir := filepath.Join(includeDir, MODULE_CODE)
	os.MkdirAll(moduleDir, os.ModePerm)

	var includeFiles []string

	var headersConfigs []HeaderConfig
	var metadataConfigs []MetadataConfig


	buildAndWriteDefines(deploymentDir, includeDir, modules, &includeFiles)

	buildAndWriteTypedefs(deploymentDir, includeDir, modules, &includeFiles)

	buildAndWriteConstants(deploymentDir, includeDir, modules, &includeFiles)


	for _, oneModule := range modules {
		headersConfigs = append(headersConfigs, parseHeaderConfig(deploymentDir, oneModule))
		metadataConfigs = append(metadataConfigs, parseMetadataConfig(deploymentDir, oneModule))
	}

	buildHeadersAndWriteHeaders(headersConfigs, metadataConfigs, includeDir, &includeFiles)

	mergedParser := MergedParser{}

	for _, blockName := range modelConfig.Pipeline{

		block := getModelBlockByName(modelConfig.ProgrammableBlocks, blockName)

		switch block.Abstraction{
			case PARSER:
				buildAndWriteParserBlock(block, &mergedParser, headersConfigs, modelConfig, includeDir)
				includeFiles = append(includeFiles, path.Join(INCLUDE, block.Filename + DOT + P4))
			case VERIFY_CHK, COMPUTE_CHK:
				buildAndWriteChkBlock(block, deploymentDir, includeDir, modules)
				includeFiles = append(includeFiles, path.Join(INCLUDE, block.Filename + DOT + P4))
			case INGRESS_MAU, EGRESS_MAU:
				buildAndWriteMauBlock(block, deploymentDir, includeDir, moduleDir, modules)
				includeFiles = append(includeFiles, path.Join(INCLUDE, block.Filename + DOT + P4))
			case DEPARSER:
				buildAndWriteDeparserBlock(block, &mergedParser, includeDir)
				includeFiles = append(includeFiles, path.Join(INCLUDE, block.Filename + DOT + P4))
		}
	}

	writeMainBlock(modelConfig, pipelineDir, &includeFiles)

}

func buildAndWriteDefines(deploymentDir, includeDir string, modules []string, includes *[]string) {

	fmt.Printf("Writing defines... ")

	var spaceTimes int = 0
	defineFile := path.Join(includeDir, DEFINE + DOT  + P4)

	var defineUnion = make(map[string]Define)

	for _, module := range modules{
		defines := getDefineConfig(deploymentDir, module)
		for _, define := range defines.DefineList{
			if _, contains := defineUnion[define.Name]; !contains {
				defineUnion[define.Name] = define
			}
		}
	}

	if len(defineUnion) == 0{
		fmt.Printf("NONE EXIST\n")
		return
	}

	f, err := os.OpenFile(defineFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	spaces := strings.Repeat(SPACE, spaceTimes)
	for _, define := range defineUnion {
		line := fmt.Sprintf(SPACES + DEFINE_STATEMENT + BREAKLINE, spaces, define.Name, strconv.Itoa(define.Value))
		if _, err := f.WriteString(line); err != nil {
			log.Println(err)
		}
	}

	if err == nil {
		*includes = append(*includes, path.Join(INCLUDE, DEFINE + DOT  + P4))
		fmt.Printf(OK + BREAKLINE)
	}else{
		fmt.Printf(FAILED + BREAKLINE)
	}

}

func buildAndWriteConstants(deploymentDir, includeDir string, modules []string, includes *[]string) {

	fmt.Printf("Writing constants... ")

	var spaceTimes int = 0
	constantFile := path.Join(includeDir, CONSTANT + DOT + P4)

	var constantUnion = make(map[string]Constant)

	for _, module := range modules{
		constants := getConstantConfig(deploymentDir, module)
		for _, constant := range constants.ConstantList{
			if _, contains := constantUnion[constant.Name]; !contains {
				constantUnion[constant.Name] = constant
			}
		}
	}

	if len(constantUnion) == 0{
		fmt.Printf("NONE EXIST\n")
		return
	}

	f, err := os.OpenFile(constantFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	spaces := strings.Repeat(SPACE, spaceTimes)
	for _, constant := range constantUnion {
		statement := ""
		if isNumeric(constant.Size){
			statement = CONST_BIT_STMT
		}else{
			statement = CONST_TP_STMT

		}
		line := fmt.Sprintf(SPACES + statement + BREAKLINE, spaces, constant.Size, constant.Name,  strconv.Itoa(constant.Value))
		if _, err := f.WriteString(line); err != nil {
			log.Println(err)
		}
	}

	if err == nil {
		*includes = append(*includes, path.Join(INCLUDE, CONSTANT + DOT  + P4))
		fmt.Printf(OK + BREAKLINE)
	}else{
		fmt.Printf(FAILED + BREAKLINE)
	}

}

func buildAndWriteTypedefs(deploymentDir, includeDir string, modules []string, includes *[]string) {

	fmt.Printf("Writing typedefs... ")

	var spaceTimes int = 0
	typedefFile := path.Join(includeDir, TYPEDEF + DOT + P4)

	var typedefUnion = make(map[string]Typedef)

	for _, module := range modules{
		typedefs := getTypedefConfig(deploymentDir, module)
		for _, typedef := range typedefs.TypedefList{
			if _, contains := typedefUnion[typedef.Name]; !contains {
				typedefUnion[typedef.Name] = typedef
			}
		}
	}

	if len(typedefUnion) == 0{
		fmt.Printf("NONE EXIST\n")
		return
	}

	f, err := os.OpenFile(typedefFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	spaces := strings.Repeat(SPACE, spaceTimes)
	for _, typedef := range typedefUnion {
		line := fmt.Sprintf(SPACES + TYPEDEF_STATEMENT + BREAKLINE, spaces, strconv.Itoa(typedef.Size), typedef.Name)
		if _, err := f.WriteString(line); err != nil {
			log.Println(err)
		}
	}

	if err == nil {
		*includes = append(*includes, path.Join(INCLUDE, TYPEDEF + DOT  + P4))
		fmt.Printf(OK + BREAKLINE)
	}else{
		fmt.Printf(FAILED + BREAKLINE)
	}

}

func buildAndWriteParserBlock(block ProgrammableBlock, m *MergedParser, headersConfigs []HeaderConfig, modelConfig Model, directory string) {

	var allStates []ParsingState
	//var root string
	for _, config := range headersConfigs{
		var currentStates []ParsingState
		_, currentStates = calculateRootState(config.ParsingStates)
		//Join slices
		allStates = append(allStates, currentStates...)
		//fmt.Println("root: "+ root)
	}

	unionParseStates := getUnionParser(allStates)
	m.UnionParseStates = unionParseStates

	orderedStates := getParseStateOrder(unionParseStates)
	m.OrderedStates = orderedStates

	writeParser(block, unionParseStates, orderedStates, directory)
}

func buildAndWriteChkBlock(block ProgrammableBlock, deploymentDir, includeDir string, modules []string) {

	fmt.Printf("Writing %s... ", block.Abstraction)

	var spaceTimes int = 0
	chkFile := path.Join(includeDir, block.Filename + DOT + P4)

	writeBlockHeader(block, chkFile, spaceTimes)

	writeChkApplyBlock(block, deploymentDir, chkFile, spaceTimes, modules)

	writeEndOfBlock(chkFile, spaceTimes)

	fmt.Printf("OK \n")

}

func buildAndWriteMauBlock(block ProgrammableBlock, deploymentDir, includeDir, moduleDir string, modules []string) {

	fmt.Printf("Writing %s... ", block.Abstraction)

	var spaceTimes int = 0
	mauFile := path.Join(includeDir, block.Filename + DOT + P4)

	writeCodeBlockInclusion(block, deploymentDir, moduleDir, mauFile, modules, spaceTimes)

	writeBlockHeader(block, mauFile, spaceTimes)

	writeMauApplyBlock(block, deploymentDir, mauFile, modules, spaceTimes)

	writeEndOfBlock(mauFile, spaceTimes)

	fmt.Printf("OK \n")

}

func writeCodeBlockInclusion(block ProgrammableBlock, deploymentDir, moduleDir, file string, modules []string, spaceTimes int) {

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Writing: to include code block file
	spaces := strings.Repeat(SPACE, spaceTimes)
	line := fmt.Sprintf(SPACES, spaces)
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}

	for _, module := range modules{
		moduleCodeConfig, err := getModuleCodeConfig(deploymentDir, module)
		if err != nil{
			fmt.Println(err.Error())
			continue
		}

		for _, integration := range moduleCodeConfig.Integrate{
			if integration.Block != block.Name{
				continue
			}

			logicFile := path.Join(deploymentDir, module, CODE, integration.Logic)
			dstFile := path.Join(moduleDir, integration.Logic)

			copy(logicFile, dstFile)

			if integration.CallControl{
				// Writing: apply {
				spaces = strings.Repeat(SPACE, spaceTimes)
				line = fmt.Sprintf(SPACES + INCLUDE_STATEMENT + BREAKLINE, spaces, path.Join(MODULE_CODE, integration.Logic))
				if _, err := f.WriteString(line); err != nil {
					log.Println(err)
				}
			}
		}
	}

	// 2 breaklines
	line = fmt.Sprintf(strings.Repeat(BREAKLINE, 2))
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}

}

func buildAndWriteDeparserBlock(block ProgrammableBlock, m *MergedParser, directory string) {

	fmt.Printf("Writing %s... ", block.Abstraction)

	var spaceTimes int = 0
	deparserFile := path.Join(directory, block.Filename + DOT + P4)

	writeBlockHeader(block, deparserFile, spaceTimes)

	writeDeparserApplyBlock(m, deparserFile, spaceTimes)

	writeEndOfBlock(deparserFile, spaceTimes)

	fmt.Printf("OK \n")

}

func writeMainBlock(config Model, directory string, includes *[]string) {
	fmt.Printf("Writing main file... ")

	var spaceTimes int = 0
	mainFile := path.Join(directory, config.Main.Filename + DOT + P4)

	f, err := os.OpenFile(mainFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Writing default includes
	spaces := strings.Repeat(SPACE, spaceTimes)
	for _, defInclude := range config.DefaultLibraries{
		line := fmt.Sprintf(SPACES + DEF_INCLUDE, spaces, defInclude)
		if _, err := f.WriteString(line + NEWLINE); err != nil {
			log.Println(err)
		}
	}

	// 2 breaklines
	line := fmt.Sprintf(strings.Repeat(BREAKLINE, 2))
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}

	// Writing includes for each file created
	for _, include := range *includes{
		line := fmt.Sprintf(SPACES + INCLUDE_STATEMENT, spaces, include)
		if _, err := f.WriteString(line + NEWLINE); err != nil {
			log.Println(err)
		}
	}

	// 2 breaklines
	line = fmt.Sprintf(strings.Repeat(BREAKLINE, 2))
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}

	// Writing switch name and parenthesis open
	line = fmt.Sprintf(SPACES + SWITCH_HEADER + BREAKLINE, spaces, config.SwitchName)
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}

	spaceTimes += 4
	spaces = strings.Repeat(SPACE, spaceTimes)
	for i, block := range config.ProgrammableBlocks {
		line = fmt.Sprintf(SPACES + PARAM_SWITCH, spaces, block.Name)
		if i != len(config.ProgrammableBlocks) -1 {
			line += COMMA
		}
		line += BREAKLINE
		if _, err := f.WriteString(line); err != nil {
			log.Println(err)
		}
	}

	spaceTimes -= 4
	spaces = strings.Repeat(SPACE, spaceTimes)

	// Writing end of switch definition
	line = fmt.Sprintf(SPACES + CLOSING_SWITCH + BREAKLINE, spaces, config.Main.Filename)
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}

	if err == nil {
		fmt.Printf(OK + BREAKLINE)
	}else{
		fmt.Printf(FAILED + BREAKLINE)
	}
}

/*--------------------------------------------------------------------------------------------------------------------------------------------*/

func writeChkApplyBlock(block ProgrammableBlock, deploymentDir, file string, spaceTimes int, modules []string) {

	spaceTimes += 4

	var spaces string

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()


	spaces = strings.Repeat(SPACE, spaceTimes)
	line := fmt.Sprintf(SPACES + APPLY_OPEN + BREAKLINE, spaces)
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}

	for _, module := range modules{

		moduleCodeConfig, err := getModuleCodeConfig(deploymentDir, module)
		if err != nil{
			fmt.Println(err.Error())
			continue
		}

		for _, integration := range moduleCodeConfig.Integrate{

			if integration.Block != block.Name{
				continue
			}

			integrationFile := path.Join(deploymentDir, module, CODE, integration.Logic)

			if integration.Merge{
				writeSectionToMerge(file, integrationFile, spaceTimes);
			}
		}
	}

	spaces = strings.Repeat(SPACE, spaceTimes)
	line = fmt.Sprintf(SPACES + CLOSING_BRACKET + BREAKLINE, spaces)
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}
}

func writeMauApplyBlock(block ProgrammableBlock, deploymentDir, file string, modules []string, spaceTimes int) {

	spaceTimes += 4

	var spaces string

	for _, module := range modules{
		moduleCodeConfig, err := getModuleCodeConfig(deploymentDir, module)
		if err != nil{
			fmt.Println(err.Error())
			continue
		}

		for _, integration := range moduleCodeConfig.Integrate{
			if integration.Block != block.Name{
				continue
			}

			//logicFile := path.Join(directory, module, CODE, integration.Logic)

			if integration.CallControl{
				writeLogicDeclaration(integration, file, spaceTimes);
			}
		}

	}

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	// Writing: "several breaklines), readability
	spaces = strings.Repeat(SPACE, spaceTimes)
	line := fmt.Sprintf(SPACES + strings.Repeat(BREAKLINE, 2), spaces)
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}

	// Writing: apply {
	spaces = strings.Repeat(SPACE, spaceTimes)
	line = fmt.Sprintf(SPACES + APPLY_OPEN + BREAKLINE, spaces)
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}

	for _, module := range modules{
		moduleCodeConfig, err := getModuleCodeConfig(deploymentDir, module)
		if err != nil{
			fmt.Println(err.Error())
			continue
		}

		for _, integration := range moduleCodeConfig.Integrate{
			if integration.Block != block.Name{
				continue
			}

			//logicFile := path.Join(directory, module, CODE, integration.Logic)

			if integration.CallControl{
				writeLogicCalls(integration, file, spaceTimes);
			}
		}
	}

	spaces = strings.Repeat(SPACE, spaceTimes)
	line = fmt.Sprintf(SPACES + CLOSING_BRACKET + BREAKLINE, spaces)
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}
}

func writeDeparserApplyBlock(m *MergedParser, file string, spaceTimes int) {

	spaceTimes += 4

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()


	// Wririntg: apply {
	spaces := strings.Repeat(SPACE, spaceTimes)
	line := fmt.Sprintf(SPACES + APPLY_OPEN + BREAKLINE, spaces)
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}

	spaceTimes += 4

	for _, states := range m.OrderedStates{
		for _, state := range states{
			parsingState := m.UnionParseStates[state]

			if parsingState.Extract == ""{
				continue
			}

			// Writing: apply {
			spaces := strings.Repeat(SPACE, spaceTimes)
			line := fmt.Sprintf(SPACES + EMIT_STATEMENT + BREAKLINE, spaces, parsingState.Extract)
			if _, err := f.WriteString(line); err != nil {
				log.Println(err)
			}
		}

	}

	spaceTimes -= 4
	spaces = strings.Repeat(SPACE, spaceTimes)
	line = fmt.Sprintf(SPACES + CLOSING_BRACKET + BREAKLINE, spaces)
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}

}

func writeLogicCalls(integration Integrate, file string, spaceTimes int) {

	spaceTimes += 4
	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	spaces := strings.Repeat(SPACE, spaceTimes)

	var args []string

	for _, arg := range integration.Arguments{
		args = append(args, arg.Name)
		//TODO: check arg.Type to match block parameters?
	}


	line := fmt.Sprintf(SPACES + CONTROL_APPLY + BREAKLINE,
				spaces,
				strings.ToLower(integration.ControlName),
				strings.Join(args , ", "))

	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}


}

func writeLogicDeclaration(integration Integrate, file string, spaceTimes int) {

	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	spaces := strings.Repeat(SPACE, spaceTimes)

	line := fmt.Sprintf(SPACES + CONTROL_DEC + BREAKLINE,
							spaces, integration.ControlName, strings.ToLower(integration.ControlName))
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}
}

func writeSectionToMerge(writeFile, readFile string, spaceTimes int) {

	if !fileExists(readFile) || !fileExists(writeFile) {
		return
	}

	spaceTimes += 4
	spaces := strings.Repeat(SPACE, spaceTimes)

	rFile, err := os.Open(readFile)
	if err != nil {
		log.Fatal(err)
	}
	defer rFile.Close()

	// If the file doesn't exist, create it, or append to the file
	wFile, err := os.OpenFile(writeFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer wFile.Close()

	scanner := bufio.NewScanner(rFile)
	for scanner.Scan() {
		line := fmt.Sprintf(SPACES + scanner.Text() + BREAKLINE, spaces)
		if _, err := wFile.WriteString(line); err != nil {
			log.Println(err)
		}
	}
}

func getModuleCodeConfig(directory, module string) (ModuleCodeConfig, error){

	moduleFile := strings.Join( []string{module, CONFIG, YAML} , ".")
	fPath:= path.Join(directory, module, CODE, moduleFile)

	var err error

	if !fileExists(fPath){
		return ModuleCodeConfig{}, errors.New(fmt.Sprintf("File (%s) does not exist", moduleFile))
	}

	yamlFile, err := ioutil.ReadFile(fPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	moduleConfig :=  ModuleCodeConfig{}

	err = yaml.Unmarshal([]byte(yamlFile), &moduleConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return moduleConfig, nil
}

func buildHeadersAndWriteHeaders(headersConfigs []HeaderConfig, metadataConfigs []MetadataConfig, directory string, externalFiles *[]string) (map[string]string, map[string][]Field) {

	headers := make(map[string]string)
	metadatas := make(map[string][]Field)
	var headerTypes []string

	fmt.Printf("Writing headers... \n")

	longestHeaderType := 0
	// In this loops we check heaer type declaration and header instantiation
	for _, oneHeaderconfig := range headersConfigs {
		for _, oneType := range oneHeaderconfig.Headers {

			headerType := strings.ToLower(oneType.HeaderType) + TYPE_DEC

			if longestHeaderType < len(headerType) {
				longestHeaderType = len(headerType)
			}

			for _, declareWithName := range oneType.Statements {
				_, contains := headers[declareWithName]
				if !contains {
					headers[declareWithName] = headerType
				}
			}

			if !Contains(headerType, headerTypes) {
				fmt.Printf("%s Writing header type: %s\n", "  ", headerType)
				writeOneHeader(oneType, directory, headerType)
				headerTypes = append(headerTypes, headerType)
			}
		}
	}


	longestMetaType := 0

	for _, oneMetadataConfig := range metadataConfigs {
		for _, oneMeta := range oneMetadataConfig.Metadatas {

			_, contains := metadatas[oneMeta.Name]
			if !contains {
				metadatas[oneMeta.Name] = []Field{}
			}

			fieldList := metadatas[oneMeta.Name]

			for _, oneField := range oneMeta.Fields {
				if !ContainsField(oneField, fieldList){
					fieldList = append(fieldList, oneField)
					metadatas[oneMeta.Name] = fieldList
				}
			}

		}
	}

	writeMetaDeclaration(metadatas, directory, longestMetaType)

	if len(headers) > 0{
		*externalFiles = append(*externalFiles, path.Join(INCLUDE, HEADER_FILE))
	}


	writeHeaderDeclaration(headers, headerTypes, directory, longestHeaderType)

	return headers, metadatas
}

func getParseStateOrder(unionParseStates map[string]ParsingState) map[int][]string {

	parseStateOrder := make(map[int][]string)

	var parSingStateSlice []ParsingState
	for _, v := range unionParseStates{
		parSingStateSlice = append(parSingStateSlice, v)
	}

	var currentStates []string
	root, _ :=  calculateRootState(parSingStateSlice)
	currentStates = append(currentStates, root)
	var nextStates []string

	index := 0
	for process := true; process; process = (len(currentStates) > 0) {
		var stateGroup []string
		for i , state := range currentStates{

			if fState, found := unionParseStates[state]; found {

				if fState.NextStates != nil && len(fState.NextStates) > 0 {
					stateNames := getParseStateNames(unionParseStates[state].NextStates)
					nextStates = append(nextStates, stateNames...)

				}else {
					nextStates = append(nextStates, fState.Default.Name)
				}

				stateGroup = append(stateGroup, state)
			}

			if i == len(currentStates)-1{
				if len(stateGroup) > 0{
					parseStateOrder[index] = stateGroup
				}
				index++
				currentStates = nextStates
				nextStates = []string{}
			}
		}
	}

	return parseStateOrder

}

func getParseStateNames(states []NextState) []string {

	var names []string
	for _, state := range states{
		names = append(names, state.Name)
	}

	return names
}

func writeOneMeta(name string, fields []Field, directory string) {

	var spacetimes int = 0

	// We also include metadata for now in header file
	headerFile := path.Join(directory, HEADER_FILE)

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(headerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	spaces := strings.Repeat(SPACE, spacetimes)
	line := fmt.Sprintf(SPACES + STRUCT_META, spaces, name+TYPE_DEC)
	if _, err := f.WriteString(line + NEWLINE); err != nil {
		log.Println(err)
	}

	spacetimes += 4
	spaces = strings.Repeat(SPACE, spacetimes)

	longestType := 0

	for _, oneField := range fields {
		if _, err := strconv.Atoi(oneField.Bitwidth); err == nil {
			if longestType < len(oneField.Bitwidth) + 5{
				longestType = len(oneField.Bitwidth) + 5
			}
		}else{
			if longestType < len(oneField.Bitwidth) + 5{
				longestType = len(oneField.Bitwidth)
			}
		}
	}
	longestType += 1

	gap := 0
	for _, oneField := range fields {

		var fieldType string
		if _, err := strconv.Atoi(oneField.Bitwidth); err == nil {
			fieldType = fmt.Sprintf(BIT_TYPE, oneField.Bitwidth)
			gap = longestType - (len(oneField.Bitwidth) + 5)
		}else{
			fieldType = fmt.Sprintf(TYPEDEF_TYPE, oneField.Bitwidth)
			gap = longestType - len(oneField.Bitwidth)
		}

		spaceGap := strings.Repeat(SPACE, gap)

		line = fmt.Sprintf(SPACES + HEADER_FIELD_LINE, spaces, fieldType, spaceGap, oneField.Name)
		if _, err := f.WriteString(line + NEWLINE); err != nil {
			log.Println(err)
		}
	}

	spacetimes = 0
	spaces = strings.Repeat(SPACE, spacetimes)
	line = fmt.Sprintf(SPACES + CLOSING_BRACKET, spaces)
	if _, err := f.WriteString(CLOSING_BRACKET + strings.Repeat(NEWLINE, 2)); err != nil {
		log.Println(err)
	}

}

func writeOneHeader(oneHeader interface{}, directory, headerType string) {

	var spacetimes int = 0

	headerFile := path.Join(directory, HEADER_FILE)

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(headerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	t, ok := oneHeader.(Header) //One header with name, fields, etc.
	if !ok {

	}

	spaces := strings.Repeat(SPACE, spacetimes)

	if (t.Annotation != Annotation{}){
		line := fmt.Sprintf(SPACES + ANNOTATION_LINE, spaces, t.Annotation.Type, t.Annotation.Name)
		if _, err := f.WriteString(line + NEWLINE); err != nil {
			log.Println(err)
		}
	}

	line := fmt.Sprintf(SPACES + HEADER_LINE, spaces, headerType, OPEN_BRACKET)
	if _, err := f.WriteString(line + NEWLINE); err != nil {
		log.Println(err)
	}

	spacetimes += 4
	spaces = strings.Repeat(SPACE, spacetimes)

	longestType := 0

	for _, oneField := range t.Fields {
		if _, err := strconv.Atoi(oneField.Bitwidth); err == nil {
			if longestType < len(oneField.Bitwidth) + 5{
				longestType = len(oneField.Bitwidth) + 5
			}
		}else{
			if longestType < len(oneField.Bitwidth) + 5{
				longestType = len(oneField.Bitwidth)
			}
		}
	}
	longestType += 1

	gap := 0
	for _, oneField := range t.Fields {

		var fieldType string
		if _, err := strconv.Atoi(oneField.Bitwidth); err == nil {
			fieldType = fmt.Sprintf(BIT_TYPE, oneField.Bitwidth)
			gap = longestType - (len(oneField.Bitwidth) + 5)
		}else{
			fieldType = fmt.Sprintf(TYPEDEF_TYPE, oneField.Bitwidth)
			gap = longestType - len(oneField.Bitwidth)
		}

		spaceGap := strings.Repeat(SPACE, gap)

		line = fmt.Sprintf(SPACES + HEADER_FIELD_LINE, spaces, fieldType, spaceGap, oneField.Name)
		if _, err := f.WriteString(line + NEWLINE); err != nil {
			log.Println(err)
		}
	}

	spacetimes = 0
	spaces = strings.Repeat(SPACE, spacetimes)
	line = fmt.Sprintf(SPACES + CLOSING_BRACKET, spaces)
	if _, err := f.WriteString(CLOSING_BRACKET + strings.Repeat(NEWLINE, 2)); err != nil {
		log.Println(err)
	}

}

func writeMetaDeclaration(metas map[string][]Field, directory string, longestHeaderType int) {

	var spacetimes int = 0
	longestMetaName := 0

	headerFile := path.Join(directory, HEADER_FILE)

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(headerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for name, fields := range metas {
		writeOneMeta(name, fields, directory)

		longestMetaName = 0

		if longestMetaName < len(name) + 5{
			longestMetaName = len(name)
		}
		longestMetaName += 1
	}

	spaces := strings.Repeat(SPACE, spacetimes)
	line := fmt.Sprintf(SPACES + STRUCT_META, spaces, LOCAL_METADATA+TYPE_DEC)
	if _, err := f.WriteString(line + NEWLINE); err != nil {
		log.Println(err)
	}

	spacetimes += 4
	spaces = strings.Repeat(SPACE, spacetimes)

	if len(metas) > 0{
		fmt.Printf("%s Declaring metadata: ", "  ")
	}

	for name, _ := range metas {
		fmt.Printf("%s ", name)
		gap := longestMetaName + 1- len(name)
		spaceGap := strings.Repeat(SPACE, gap)
		line = fmt.Sprintf(SPACES + ONE_META_STMT, spaces, name+TYPE_DEC, spaceGap, name)
		if _, err := f.WriteString(line + NEWLINE); err != nil {
			log.Println(err)
		}
	}
	fmt.Printf("\n")

	spacetimes = 0
	spaces = strings.Repeat(SPACE, spacetimes)
	line = fmt.Sprintf(SPACES + CLOSING_BRACKET, spaces)
	if _, err := f.WriteString(CLOSING_BRACKET +  strings.Repeat(NEWLINE, 2)); err != nil {
		log.Println(err)
	}
}

func writeHeaderDeclaration(headers map[string]string, types []string, directory string, longestHeaderType int) {

	var spacetimes int = 0

	headerFile := path.Join(directory, HEADER_FILE)

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(headerFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	spaces := strings.Repeat(SPACE, spacetimes)
	line := fmt.Sprintf(SPACES + STRUCT_HEADERS, spaces, HEADERS+TYPE_DEC)
	if _, err := f.WriteString(line + NEWLINE); err != nil {
		log.Println(err)
	}

	spacetimes += 4
	spaces = strings.Repeat(SPACE, spacetimes)


	fmt.Printf("%s Declaring headers: ", "  ")

	for _, listedType := range types {
		for headerName, headerType := range headers {
			if listedType != headerType {
				continue;
			}
			fmt.Printf("%s ", headerName)
			gap := longestHeaderType + 3 - len(headerType)
			spaceGap := strings.Repeat(SPACE, gap)
			line = fmt.Sprintf(SPACES + ONE_HEADER_STMT, spaces, headerType, spaceGap, headerName)
			if _, err := f.WriteString(line + NEWLINE); err != nil {
				log.Println(err)
			}
		}
	}
	fmt.Printf("\n")

	spacetimes = 0
	spaces = strings.Repeat(SPACE, spacetimes)
	line = fmt.Sprintf(SPACES + CLOSING_BRACKET, spaces)
	if _, err := f.WriteString(CLOSING_BRACKET +  strings.Repeat(NEWLINE, 2)); err != nil {
		log.Println(err)
	}
}

func writeParser(block ProgrammableBlock, parseStates map[string]ParsingState,
	orderedStateNames map[int][]string, directory string){

	fmt.Sprintf("Writing %s...", block.Abstraction)

	var spaceTimes int = 0
	parserFile := path.Join(directory, PARSER_FILE)

	writeBlockHeader(block, parserFile, spaceTimes)

	for _, groupStates := range orderedStateNames{
		writeParserStates(parseStates , groupStates, parserFile)
	}

	writeEndOfBlock(parserFile, spaceTimes)

	fmt.Sprintf("OK \n")

}

func writeParserStates(statesConfig map[string]ParsingState, statesNames []string, parserFile string) {

	var spaceTimes = 4
	var spaces string

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(parserFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	for _, stateName := range statesNames{

		stateConfig := statesConfig[stateName]

		spaces = strings.Repeat(SPACE, spaceTimes)
		//writing: state parse_whatever {
		line := fmt.Sprintf(SPACES + PARSER_STATE + NEWLINE, spaces, STATE, stateName)
		if _, err := f.WriteString(line); err != nil {
			log.Println(err)
		}

		spaceTimes += 4
		spaces = strings.Repeat(SPACE, spaceTimes)

		if stateConfig.Extract != ""{
			line = fmt.Sprintf(SPACES + EXTRACT_STATEMENT + NEWLINE, spaces, "packet", "hdr", stateConfig.Extract)
			if _, err := f.WriteString(line); err != nil {
				log.Println(err)
			}
		}

		if stateConfig.NextStates == nil || len(stateConfig.NextStates) == 0{
			//Writing: transition toSomeState;
			line = fmt.Sprintf(SPACES + SIMPLE_TRANSITION + NEWLINE, spaces, stateConfig.Default.Name)
			if _, err := f.WriteString(line); err != nil {
				log.Println(err)
			}
		}else {
			transCondition := ""

			if stateConfig.IsMeta {
				transCondition = fmt.Sprintf(TRANSITION_META, stateConfig.OnHeader, stateConfig.OnField)
			}else {
				transCondition = fmt.Sprintf(TRANSITION_COND, "hdr", stateConfig.OnHeader, stateConfig.OnField)
			}
			//writing: transition select(onSomeField) {
			line = fmt.Sprintf(SPACES + TRANSITION_FUNC + NEWLINE, spaces,  transCondition)
			if _, err := f.WriteString(line); err != nil {
				log.Println(err)
			}

			line = ""
			spaceTimes += 4
			spaces = strings.Repeat(SPACE, spaceTimes)
			for i, nextState := range stateConfig.NextStates{
				//writing: onValue: toSomeState;
				line = fmt.Sprintf(SPACES + ONE_SELECT_TR + NEWLINE, spaces, nextState.OnValue, nextState.Name)
				if _, err := f.WriteString(line); err != nil {
					log.Println(err)
				}

				if i == len(stateConfig.NextStates) - 1 {
					line = fmt.Sprintf(SPACES + ONE_SELECT_TR + NEWLINE, spaces, DEFAULT, ACCEPT)
					if _, err := f.WriteString(line); err != nil {
						log.Println(err)
					}
				}

			}

			spaceTimes -= 4
			spaces = strings.Repeat(SPACE, spaceTimes)
			// Writing: }
			line = fmt.Sprintf(SPACES + CLOSING_BRACKET + NEWLINE, spaces)
			if _, err := f.WriteString(line); err != nil {
				log.Println(err)
			}
		}

		spaceTimes -= 4
		spaces = strings.Repeat(SPACE, spaceTimes)
		// Writing: }
		line = fmt.Sprintf(SPACES + CLOSING_BRACKET, spaces)
		if _, err := f.WriteString(line); err != nil {
			log.Println(err)
		}

		if _, err := f.WriteString(strings.Repeat(NEWLINE, 2)); err != nil {
			log.Println(err)
		}
	}

}

func writeBlockHeader(block ProgrammableBlock, parserFile string, spaceTimes int) {

	headerIndentSpaces := len(block.Type) + len(SPACE) + len(block.Name) + len(SPACE) + len(OPEN_PARENTHESIS)
	arguments := generateArguments(block, headerIndentSpaces)

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(parserFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()

	spaces := strings.Repeat(SPACE, spaceTimes)
	newLines := strings.Repeat(NEWLINE, 2)
	line := fmt.Sprintf(SPACES + BLOCK_FUNC_HEADER + newLines, spaces, block.Type, block.Name, arguments)
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}
}

func writeEndOfBlock(file string, spacetimes int) {

	var spaces string

	// If the file doesn't exist, create it, or append to the file
	f, err := os.OpenFile(file, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()


	spaces = strings.Repeat(SPACE, spacetimes)
	line := fmt.Sprintf(SPACES + CLOSING_BRACKET + strings.Repeat(NEWLINE, 2), spaces)
	if _, err := f.WriteString(line); err != nil {
		log.Println(err)
	}

}

func generateArguments(block ProgrammableBlock, indentSpaces int) string {

	var completeArgs string
	var argumentStr string

	spaceTimes := 0
	spaces := ""
	for i, parameter := range block.Parameters{

		if i != 0{
			spaceTimes = indentSpaces
			spaces = strings.Repeat(SPACE, spaceTimes)

		}

		if parameter.Direction == ""{
			argumentStr = fmt.Sprintf(SPACES + ARGUMENT_2, spaces, parameter.Type, parameter.Name)
		}else{
			argumentStr = fmt.Sprintf(SPACES + ARGUMENT_3, spaces, parameter.Direction, parameter.Type, parameter.Name)
		}

		if i != len(block.Parameters) -1 {
			argumentStr += COMMA + BREAKLINE
		}

		completeArgs += argumentStr
	}

	return completeArgs
}

func getModelBlockByName(blocks []ProgrammableBlock, name string) ProgrammableBlock{

	var foundBlock ProgrammableBlock
	for _, oneBlock := range blocks{
		if oneBlock.Name == name{
			foundBlock = oneBlock
			return foundBlock
		}
	}
	return foundBlock
}

func getUnionParser(allStates []ParsingState)(map[string]ParsingState) {

	var unionParseStates = make(map[string]ParsingState)

	for _, oneState := range allStates{

		if state, contains := unionParseStates[oneState.Name]; !contains {
			unionParseStates[oneState.Name] = oneState
			continue
		}else{
			//For now, check names are the same (ie.e. state variable) AND ALSO that extract the same headers
			if state.Extract == oneState.Extract{
				for _, tr := range oneState.NextStates{
					if !containsTransition(tr, state.NextStates){
						if state.NextStates == nil {
							state.NextStates = []NextState{}
							if state.Transition == DIRECT{
								state.Transition = CONDITIONAL
							}
							if state.Default.Name != ACCEPT {
								//fmt.Printf("State %s direct transition to %s (non accept) state.\n", state.Name, state.Default.Name)
								state.Default.Name = ACCEPT
							}

							if state.OnHeader == ""{
								state.OnHeader = oneState.OnHeader
								if oneState.IsMeta {
									state.IsMeta = oneState.IsMeta
								}
							}else{
								if state.OnHeader != oneState.OnHeader{
									fmt.Println("Warning:  Adding a transition to a state that has different header to decide next state")
								}
							}


							if state.OnField == ""{
								state.OnField = oneState.OnField
							}else{
								if state.OnField != oneState.OnField{
									fmt.Println("Warning:  Adding a transition to a state that has different field to decide next state")
								}
							}




						}
						state.NextStates = append(state.NextStates, tr)
					}
				}
				unionParseStates[oneState.Name] = state
			}else {
				log.Println("Parsing state names are the same but not the header extraction.")
			}
		}
	}

	return unionParseStates
}

func containsTransition(tr NextState, states []NextState) bool {

	for _, nextState := range states{
		if nextState.Name == tr.Name{
			return true
		}
	}
	return false
}

func PullConfiguration(modules []string, model, target, depDir, repoDomain string){

	fmt.Println(fmt.Sprintf("%s", repoDomain))
	if(strings.HasPrefix(repoDomain, "/")){
		//This is a local domain


	}else{

		u, err := url.Parse(HTTPS_PROTO+repoDomain)
		if err != nil {
			panic("invalid url")
		}

		modelFile := fmt.Sprintf(MODEL_CONFIG_FILE, model)
		u.Path = path.Join(u.Path, target, model, modelFile)
		fPath := path.Join(depDir, modelFile)

		start := time.Now()

		fmt.Printf("Downloading %s model configuration file... ", model)
		DownloadFile(u.String(), fPath)

		duration := time.Since(start)
		fmt.Println(fmt.Sprintf("Model download time: %d", duration.Microseconds()))


		u.Path = ""

		for i := 0; i < len(modules); i++ {
			u.Path = path.Join(u.Path, target, model, modules[i]+P4Z)
			fPath := path.Join(depDir, modules[i]+P4Z)

			start := time.Now()

			fmt.Printf("Downloading module %s (%s/%s)... ", modules[i], target, model)
			DownloadFile(u.String(), fPath)

			duration := time.Since(start)
			fmt.Println(fmt.Sprintf("Download time for module %s: %d", modules[i], duration.Microseconds()))

			start = time.Now()

			fmt.Printf("Decompressing module %s (%s/%s)... ", modules[i], target, model)
			localFileDir := path.Join(depDir, modules[i]+P4Z)
			Unzip(localFileDir, depDir)

			duration = time.Since(start)
			fmt.Println(fmt.Sprintf("Decompress time for %s: %d", modules[i], duration.Microseconds()))

			u.Path = ""
		}
	}
}

func parseHeaderConfig(depDir,module string) HeaderConfig {

	fPath:= path.Join(depDir, module, DEPENDENCIES, HEADER, HEADER_DEP_FILE)

	yamlFile, err := ioutil.ReadFile(fPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	oneHeader := HeaderConfig{}

	err = yaml.Unmarshal([]byte(yamlFile), &oneHeader)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return oneHeader
}

func parseMetadataConfig(depDir,module string) MetadataConfig {

	fPath:= path.Join(depDir, module, DEPENDENCIES, METADATA, METADATA_DEP_FILE)

	yamlFile, err := ioutil.ReadFile(fPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	oneMeta := MetadataConfig{}

	err = yaml.Unmarshal([]byte(yamlFile), &oneMeta)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return oneMeta
}

func parseModelConfiguration(depDir, model string) Model {

	modelFile := fmt.Sprintf(MODEL_CONFIG_FILE, model)
	fPath := path.Join(depDir, modelFile)

	yamlFile, err := ioutil.ReadFile(fPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	modelConfig := Model{}

	err = yaml.Unmarshal([]byte(yamlFile), &modelConfig)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return modelConfig

}

func getDefineConfig(directory string, module string) Defines {

	fPath:= path.Join(directory, module, DEPENDENCIES, DEFINE, DEFINE_DEP_FILE)

	if !fileExists(fPath){
		return Defines{}
	}

	yamlFile, err := ioutil.ReadFile(fPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	defines := Defines{}

	err = yaml.Unmarshal([]byte(yamlFile), &defines)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return defines
}

func getConstantConfig(directory string, module string) Constants {

	fPath:= path.Join(directory, module, DEPENDENCIES, CONSTANT, CONSTANT_DEP_FILE)

	if !fileExists(fPath){
		return Constants{}
	}

	yamlFile, err := ioutil.ReadFile(fPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	constants := Constants{}

	err = yaml.Unmarshal([]byte(yamlFile), &constants)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return constants
}

func getTypedefConfig(directory string, module string) Typedefs {

	fPath:= path.Join(directory, module, DEPENDENCIES, TYPEDEF, TYPEDEF_DEP_FILE)

	if !fileExists(fPath){
		return Typedefs{}
	}

	yamlFile, err := ioutil.ReadFile(fPath)
	if err != nil {
		log.Printf("yamlFile.Get err   #%v ", err)
	}

	typedefs := Typedefs{}

	err = yaml.Unmarshal([]byte(yamlFile), &typedefs)
	if err != nil {
		log.Fatalf("error: %v", err)
	}

	return typedefs
}

// https://stackoverflow.com/questions/10510691/how-to-check-whether-a-file-or-directory-exists
func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return false, err
}

// https://progolang.com/how-to-download-files-in-go/
func DownloadFile(url string, filepath string) error {
	out, err := os.Create(filepath)
	if err != nil {
		return err
	}
	defer out.Close()

	// Get the data
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	fmt.Println("OK")

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return err
	}

	return nil
}

// https://stackoverflow.com/a/24792688
func Unzip(src, dest string) error {
	//src is zip file path
	r, err := zip.OpenReader(src)
	if err != nil {
		return err
	}
	defer func() {
		if err := r.Close(); err != nil {
			panic(err)
		}
	}()

	os.MkdirAll(dest, 0755)

	// Closure to address file descriptors issue with all the deferred .Close() methods
	extractAndWriteFile := func(f *zip.File) error {
		rc, err := f.Open()
		if err != nil {
			return err
		}
		defer func() {
			if err := rc.Close(); err != nil {
				panic(err)
			}
		}()

		path := filepath.Join(dest, f.Name)

		// Check for ZipSlip (Directory traversal)
		if !strings.HasPrefix(path, filepath.Clean(dest) + string(os.PathSeparator)) {
			return fmt.Errorf("illegal file path: %s", path)
		}

		if f.FileInfo().IsDir() {
			os.MkdirAll(path, f.Mode())
		} else {
			os.MkdirAll(filepath.Dir(path), f.Mode())
			f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
			if err != nil {
				return err
			}
			defer func() {
				if err := f.Close(); err != nil {
					panic(err)
				}
			}()

			_, err = io.Copy(f, rc)
			if err != nil {
				return err
			}
		}
		return nil
	}

	for _, f := range r.File {
		err := extractAndWriteFile(f)
		if err != nil {
			return err
		}
	}

	fmt.Println("OK")

	return nil
}

func calculateRootState(parsingStates []ParsingState) (string, []ParsingState){

	var allStates []ParsingState

	//check that all "start" states have direct transition and no header extraction
	var transAcct = make(map[string]uint)
	var rootState string

	for _, oneState := range parsingStates{

		allStates = append(allStates, oneState)

		oneStateName := strings.TrimSpace(oneState.Name)
		oneStateName = strings.ToLower(oneStateName)

		//add current state to map if not present
		if _, contains := transAcct[oneStateName]; !contains {
			transAcct[oneStateName] = 0
		}

		defaultName := strings.TrimSpace(oneState.Default.Name)
		defaultName = strings.ToLower(defaultName)

		//get default state to count
		if defaultName != ""{
			if _, contains := transAcct[oneState.Default.Name]; !contains {
				transAcct[oneState.Default.Name] = 1
			}else{
				transAcct[oneState.Default.Name] += 1
			}
		}

		/*if oneState.NextStates == nil || len(oneState.NextStates) == 0 {
			// Do something with no nextStates?
		}*/

		for _, nextState := range oneState.NextStates{

			stateName := strings.TrimSpace(nextState.Name)
			stateName = strings.ToLower(stateName)

			if _, contains := transAcct[stateName]; !contains {
				transAcct[stateName] = 1
			}else{
				transAcct[stateName] += 1
			}
		}

	}

	lowest := ^uint(0)
	for name, countedState := range transAcct{
		if countedState < lowest{
			lowest = countedState
			rootState = name
		}
	}

	return rootState, allStates

}

func Contains(findElement string, elements []string) bool {
	for _, oneElement := range elements {
		if oneElement == findElement {
			return true
		}
	}
	return false
}

func ContainsField(findElement Field, elements []Field) bool {
	for _, oneElement := range elements {
		if oneElement.Name == findElement.Name &&
						oneElement.Bitwidth == findElement.Bitwidth {
			return true
		}
	}
	return false
}

func fileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func isNumeric(s string) bool {
	_, err := strconv.ParseFloat(s, 64)
	return err == nil
}

func copy(src, dst string) (int64, error) {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return 0, err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return 0, fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer destination.Close()
	nBytes, err := io.Copy(destination, source)
	return nBytes, err
}
