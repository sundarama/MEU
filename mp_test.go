// mp-test.go
package main

import (
	"testing"
)

func TestMain(m *testing.M) {
	m.Run()
}

// Message has no mention, emoticon or url
func TestMessageNothing(t *testing.T) {

	msg := "test"
	respData, err := processMsgHelper(&msg)

	if err != nil {
		t.Fatalf("Failed with an error")
	}

	if _, ok := (respData)[MentionType]; ok == true {
		t.Fatalf("Expected to have no mentions")
	}

	if _, ok := (respData)[EmoticonType]; ok == true {
		t.Fatalf("Expected to have no emoticons")
	}

	if _, ok := (respData)["url"]; ok == true {
		t.Fatalf("Expected to have no url")
	}
}

// Message has only a mention
func TestMessageWithOneMention(t *testing.T) {

	msg := "test @man"
	respData, err := processMsgHelper(&msg)

	if err != nil {
		t.Fatalf("Failed with an error")
	}

	if v, ok := (respData)[MentionType]; ok == false || len(v) != 1 {
		t.Fatalf("Expected to have 1 mention")
	}

	if _, ok := (respData)[EmoticonType]; ok == true {
		t.Fatalf("Expected to have no emoticon")
	}

	if _, ok := (respData)["url"]; ok == true {
		t.Fatalf("Expected to have no url")
	}

	var mVal *string = respData[MentionType][0].(*string)
	if *mVal != "man" {
		t.Fatalf("Expected mention value of 'man'")
	}
}

// Message has only a emoticon
func TestMessageWithOneEmoticon(t *testing.T) {

	msg := "this (thisisaoneemoti)"
	respData, err := processMsgHelper(&msg)

	if err != nil {
		t.Fatalf("Failed with an error")
	}

	if _, ok := (respData)[MentionType]; ok == true {
		t.Fatalf("Expected to have no mention")
	}

	if v, ok := (respData)[EmoticonType]; ok == false || len(v) != 1 {
		t.Fatalf("Expected to have 1 emoticon")
	}

	if _, ok := (respData)["url"]; ok == true {
		t.Fatalf("Expected to have no url")
	}

	var eVal *string = respData[EmoticonType][0].(*string)
	if *eVal != "thisisaoneemoti" {
		t.Fatalf("Expected emoticon value of 'thisisaoneemoti'")
	}

}

// Message has only a url
func TestMessageWithOneUrl(t *testing.T) {

	msg := "test http://cnn.com"
	respData, err := processMsgHelper(&msg)

	if err != nil {
		t.Fatalf("Failed with an error")
	}

	if _, ok := (respData)[MentionType]; ok == true {
		t.Fatalf("Expected to have no mention")
	}

	if _, ok := (respData)[EmoticonType]; ok == true {
		t.Fatalf("Expected to have no emoticonss")
	}

	if v, ok := (respData)[UrlType]; ok == false || len(v) != 1 {
		t.Fatalf("Expected to have 1 url")
	}

	var uInfo *urlInfo = respData[UrlType][0].(*urlInfo)
	if len(uInfo.Title) == 0 {
		t.Fatalf("Expected title for the url")
	}
}

// Message starts with a mention, and no emoticon or url. Regex for mention
// has a condition that mention should start with an @ at the start of the word.
// So start of a line or message should also be valid
func TestMessageWithMentionStart(t *testing.T) {

	msg := "@man test"
	respData, err := processMsgHelper(&msg)

	if err != nil {
		t.Fatalf("Failed with an error")
	}

	if v, ok := (respData)[MentionType]; ok == false || len(v) != 1 {
		t.Fatalf("Expected to have 1 mention")
	}

	if _, ok := (respData)[EmoticonType]; ok == true {
		t.Fatalf("Expected to have no emoticon")
	}

	if _, ok := (respData)["url"]; ok == true {
		t.Fatalf("Expected to have no url")
	}

	var mVal *string = respData[MentionType][0].(*string)
	if *mVal != "man" {
		t.Fatalf("Expected mention value of 'man'")
	}
}

// Message with 1 mention, 1 emoticon, 1 url
func TestMessageWithOneEach(t *testing.T) {

	msg := "test     @man (emoticonnnnnnnn) http://cnn.com"
	respData, err := processMsgHelper(&msg)

	if err != nil {
		t.Fatalf("Failed with an error")
	}

	if v, ok := (respData)[MentionType]; ok == false || len(v) != 1 {
		t.Fatalf("Expected to have no mention")
	}

	if v, ok := (respData)[EmoticonType]; ok == false || len(v) != 1 {
		t.Fatalf("Expected to have no emoticonss")
	}

	if v, ok := (respData)[UrlType]; ok == false || len(v) != 1 {
		t.Fatalf("Expected to have 1 url")
	}

	var mention *string = respData[MentionType][0].(*string)
	if *mention != "man" {
		t.Fatalf("Mismatched mention %s", *mention)
	}

	var emoticon *string = respData[EmoticonType][0].(*string)
	if *emoticon != "emoticonnnnnnnn" {
		t.Fatalf("Mismatched emoticon")
	}

	var uInfo *urlInfo = respData[UrlType][0].(*urlInfo)
	if uInfo.Url != "http://cnn.com" {
		t.Fatalf("Expected cnn.com for url")
	}

	if len(uInfo.Title) == 0 {
		t.Fatalf("Expected title for the url")
	}
}

// Message that is funky with some wierd cases
func TestMessageFunky(t *testing.T) {

	msg := "test @(emoticonnnnnnnn) @man (Zebrahttp://cnn.com"
	respData, err := processMsgHelper(&msg)

	if err != nil {
		t.Fatalf("Failed with an error")
	}

	if v, ok := (respData)[MentionType]; ok == false || len(v) != 1 {
		t.Fatalf("Expected to have no mention")
	}

	if v, ok := (respData)[EmoticonType]; ok == false || len(v) != 1 {
		t.Fatalf("Expected to have no emoticonss")
	}

	if v, ok := (respData)[UrlType]; ok == false || len(v) != 1 {
		t.Fatalf("Expected to have 1 url")
	}

	var mention *string = respData[MentionType][0].(*string)
	if *mention != "man" {
		t.Fatalf("Mismatched mention")
	}

	var emoticon *string = respData[EmoticonType][0].(*string)
	if *emoticon != "emoticonnnnnnnn" {
		t.Fatalf("Mismatched emoticon")
	}

	var uInfo *urlInfo = respData[UrlType][0].(*urlInfo)
	if uInfo.Url != "http://cnn.com" {
		t.Fatalf("Expected cnn.com for url")
	}

	if len(uInfo.Title) == 0 {
		t.Fatalf("Expected title for the url")
	}
}

// No 'mention' because the mention word does not start with @
func TestMessageWithErrMention(t *testing.T) {

	msg := "test kk@man"
	respData, err := processMsgHelper(&msg)

	if err != nil {
		t.Fatalf("Failed with an error")
	}

	if _, ok := (respData)[MentionType]; ok == true {
		t.Fatalf("Expected to have no mention")
	}

	if _, ok := (respData)[EmoticonType]; ok == true {
		t.Fatalf("Expected to have no emoticon")
	}

	if _, ok := (respData)[UrlType]; ok == true {
		t.Fatalf("Expected to have no url")
	}
}

// No 'emoticon' because it is equal to 15 chars
func TestMessageWithErrEmoticon(t *testing.T) {

	msg := "test (thisisgolang)"
	respData, err := processMsgHelper(&msg)

	if err != nil {
		t.Fatalf("Failed with an error")
	}

	if _, ok := (respData)[MentionType]; ok == true {
		t.Fatalf("Expected to have no mention")
	}

	if _, ok := (respData)[EmoticonType]; ok == true {
		t.Fatalf("Expected to have no emoticon")
	}

	if _, ok := (respData)[UrlType]; ok == true {
		t.Fatalf("Expected to have no url")
	}
}

// No 'url' because url fetch should fail to resolve
func TestMessageWithErrUrl(t *testing.T) {

	msg := "test http://abcedefghijklmnopqrstuvwxyz1234567890.com"
	respData, err := processMsgHelper(&msg)

	if err != nil {
		t.Fatalf("Failed with an error")
	}

	if _, ok := (respData)[MentionType]; ok == true {
		t.Fatalf("Expected to have no mention")
	}

	if _, ok := (respData)[EmoticonType]; ok == true {
		t.Fatalf("Expected to have no emoticonss")
	}

	if _, ok := (respData)[UrlType]; ok == true {
		t.Fatalf("Expected to have no url")
	}

}
