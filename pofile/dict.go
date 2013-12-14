package pofile

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"io/ioutil"
	"regexp"
	"strings"
)

// import "fmt"
// import "bufio"

/*

To create a empty dictionary:

	dict := pofile.NewDictionary()
	dict := pofile.Dictionary{}

To add message:

	dict.AddMessage(msgId, msgStr)

To remove message:

	dict.RemoveMessage(msgId)

*/
var CommentRegExp = regexp.MustCompile("^\\s*#")
var EmptyLineRegExp = regexp.MustCompile("^\\s*$")
var MsgIdRegExp = regexp.MustCompile("^msgid\\s+\"(.*)\"")
var MsgStrRegExp = regexp.MustCompile("^msgstr\\s+\"(.*)\"")
var StringRegExp = regexp.MustCompile("\"(.*)\"")

type Dictionary map[string]string

func NewDictionary() *Dictionary {
	return &Dictionary{}
}

/*
Add a message with a message id
*/
func (self Dictionary) AddMessage(msgId string, msgStr string) {
	self[msgId] = msgStr
}

/*
Check if a message is defined.
*/
func (self Dictionary) HasMessage(msgId string) bool {
	_, ok := self[msgId]
	return ok
}

/*
Remove a message by a message id from the dictionary.
*/
func (self Dictionary) RemoveMessage(msgId string) {
	delete(self, msgId)
}

/*
Parse po file and load the messages into current object.
*/
func (self Dictionary) ParseAndLoadFromFile(filename string) error {
	// process(filename)
	bytes, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}
	return self.ParseAndLoad(string(bytes))
}

func (self Dictionary) ParseAndLoad(content string) error {
	lines := strings.Split(content, "\n")
	lastMsgId := []string{}
	lastMsgStr := []string{}

	state := STATE_COMPLETE

	for _, line := range lines {
		if len(line) == 0 || EmptyLineRegExp.MatchString(line) { // skip empty lines
			if state == STATE_MSGSTR {
				self.AddMessage(strings.Join(lastMsgId, ""), strings.Join(lastMsgStr, ""))
				lastMsgId = []string{}
				lastMsgStr = []string{}
				state = STATE_COMPLETE
			}
			continue
		}

		if line[0] == '#' || CommentRegExp.MatchString(line) {
			state = STATE_COMMENT
			continue
		}

		if strings.HasPrefix(line, "msgid") || MsgIdRegExp.MatchString(line) {

			state = STATE_MSGID
			msgId := MsgIdRegExp.FindStringSubmatch(line)[1]
			lastMsgId = append(lastMsgId, msgId)

		} else if strings.HasPrefix(line, "msgstr") || MsgStrRegExp.MatchString(line) {
			state = STATE_MSGSTR
			msgStr := MsgStrRegExp.FindStringSubmatch(line)[1]
			lastMsgStr = append(lastMsgStr, msgStr)
		} else if StringRegExp.MatchString(line) {
			var str = StringRegExp.FindStringSubmatch(line)[1]
			if state == STATE_MSGID {
				lastMsgId = append(lastMsgId, str)
			} else if state == STATE_MSGSTR {
				lastMsgStr = append(lastMsgStr, str)
			}
		}
	}
	return nil
}

func (self Dictionary) Merge(dict *Dictionary) {
	for key, val := range *dict {
		self[key] = val
	}
}

func (self Dictionary) MergeFile(filename string) error {
	newDict := NewDictionary()
	if err := newDict.ParseAndLoadFromFile(filename); err != nil {
		return err
	}
	self.Merge(newDict)
	return nil
}

func (self Dictionary) CSVString() string {
	var buf = bytes.NewBufferString("")
	var writer = csv.NewWriter(buf)

	writer.Write([]string{"MessageID", "MessageString"})
	for key, val := range self {
		writer.Write([]string{key, val})
	}
	writer.Flush()
	return buf.String()
}

func (self Dictionary) JSONString() string {
	jsonBytes, err := json.MarshalIndent(self, "", "  ")
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func (self Dictionary) String() string {
	return self.JSONString()
}

func (self Dictionary) WriteJSONFile(filepath string) error {
	var output = self.JSONString()
	return ioutil.WriteFile(filepath, []byte(output), 0666)
}

func (self Dictionary) WriteCSVFile(filepath string) error {
	var output = self.CSVString()
	return ioutil.WriteFile(filepath, []byte(output), 0666)
}

func (self Dictionary) LoadJSONString(content string) error {
	return self.LoadJSON([]byte(content))
}

func (self Dictionary) LoadCSVString(content string) error {
	return self.LoadCSV([]byte(content))
}

func (self Dictionary) LoadCSV(b []byte) error {
	var buf = bytes.NewBuffer(b)
	var reader = csv.NewReader(buf)
	items, err := reader.ReadAll()
	items = items[1:] // truncate first row
	for _, item := range items {
		self.AddMessage(item[0], item[1])
	}
	return err
}

func (self Dictionary) LoadJSON(b []byte) error {
	return json.Unmarshal(b, &self)
}

func (self Dictionary) LoadCSVFile(file string) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return self.LoadCSV(bytes)
}

func (self Dictionary) LoadJSONFile(file string) error {
	bytes, err := ioutil.ReadFile(file)
	if err != nil {
		return err
	}
	return self.LoadJSON(bytes)
}
