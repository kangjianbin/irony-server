package main

import (
	"fmt"
)

type TUData struct {
	tu    TranslationUnit
	file  string
	flags []string
	ref   int
}

type TUCache struct {
	index        Index
	parseOptions uint32
	tuMap        map[string]*TUData
}

func newTUData(tu TranslationUnit, file string, flags []string) *TUData {
	return &TUData{tu, file, flags, 1}
}

func (td *TUData) Dispose() {
	td.ref -= 1
	if td.ref < 0 {
		msg := fmt.Sprintf("tu for file %s, ref %d", td.file, td.ref)
		panic(msg)
	}
	if td.ref == 0 {
		td.tu.Dispose()
	}
}

func (td *TUData) Ref() {
	td.ref += 1
}

func flagsIsMatch(flags1 []string, flags2 []string) bool {
	if len(flags1) != len(flags2) {
		return false
	}
	for i := range flags1 {
		if flags1[i] != flags2[i] {
			return false
		}
	}
	return true
}

func NewTuCache() *TUCache {
	var tc TUCache

	tc.index = NewIndex(0, 0)
	tc.parseOptions = DefaultEditingTranslationUnitOptions()
	tc.parseOptions |= TranslationUnit_DetailedPreprocessingRecord |
		TranslationUnit_Incomplete | TranslationUnit_CreatePreambleOnFirstParse |
		TranslationUnit_KeepGoing | TranslationUnit_IncludeBriefCommentsInCodeCompletion
	tc.tuMap = make(map[string]*TUData)
	return &tc
}

func (tc *TUCache) Dispose() {
	tc.index.Dispose()
	for _, v := range tc.tuMap {
		v.Dispose()
	}
}

func (tc *TUCache) findTU(file string, flags []string) *TUData {
	td, ok := tc.tuMap[file]
	if !ok {
		return nil
	}
	if !flagsIsMatch(td.flags, flags) {
		delete(tc.tuMap, file)
		td.Dispose()
		return nil
	}
	return td
}

func (tc *TUCache) addTU(filename string, flags []string, tu TranslationUnit) *TUData {
	if _, ok := tc.tuMap[filename]; ok {
		exitError("BUG, tu %s already exists\n", filename)
	}
	td := newTUData(tu, filename, flags)
	tc.tuMap[filename] = td
	return td
}

func (tc *TUCache) deleteTU(filename string) {
	tu, ok := tc.tuMap[filename]
	if !ok {
		return
	}
	delete(tc.tuMap, filename)
	tu.Dispose()
}

func (tc *TUCache) Parse(filename string, flags []string, unsaved []UnsavedFile) *TUData {
	var tu TranslationUnit

	flags = append([]string{"clang"}, flags...)
	if ClangHeaderDir != "" {
		buildinFlags := []string{"-isystem", ClangHeaderDir}
		flags = append(flags, buildinFlags...)
	}
	td := tc.findTU(filename, flags)
	if td == nil {
		errCode := tc.index.ParseTranslationUnit2FullArgv(filename, flags, unsaved, tc.parseOptions, &tu)
		if !tu.IsValid() {
			logInfo("Parse failed: %d\n", errCode)
			return nil
		}
		td = tc.addTU(filename, flags, tu)
	}
	tu = td.tu
	err := tu.ReparseTranslationUnit(unsaved, tu.DefaultReparseOptions())
	if err != 0 {
		logInfo("ReParse failed, err %d\n", err)
		tc.deleteTU(filename)
		return nil
	}
	td.Ref()

	return td
}

func (tc *TUCache) GenTU(file string, flags []string, unsaved []UnsavedFile) *TUData {
	td := tc.findTU(file, flags)
	if td != nil {
		td.Ref()
		return td
	}
	return tc.Parse(file, flags, unsaved)
}
