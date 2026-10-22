package main

import (
	"time"

	"charm.land/bubbles/v2/textinput"
)

type tabType int

const (
	tabProfile tabType = iota
	tabCards
	tabData
	tabLocalData
)

func (t tabType) String() string {
	switch t {
	case tabProfile:
		return "profile"
	case tabCards:
		return "cards"
	case tabData:
		return "data"
	case tabLocalData:
		return "localdata"
	default:
		return ""
	}
}

type cardActionsType int

const (
	cardActionsUpdate cardActionsType = iota
	cardActionsDelete
)

type dataActionsType int

const (
	dataActionsDownload dataActionsType = iota
	dataActionsReload
	dataActionsDelete
)

type pageType int

const (
	landingPage pageType = iota
	registerPage
	loginPage
	confirmPage
	homePage
	profileFormPage
	cardFormPage
	filePickerPage
)

func (p pageType) String() string {
	switch p {
	case landingPage:
		return "landing"
	case registerPage:
		return "registration"
	case loginPage:
		return "login"
	case confirmPage:
		return "confirm"
	case homePage:
		return "home"
	case profileFormPage:
		return "profile/update"
	case cardFormPage:
		return "cards/create"
	case filePickerPage:
		return "files"
	default:
		return ""
	}
}

type authForm struct {
	email   textinput.Model
	confirm textinput.Model
	pass    textinput.Model
}

type cardForm struct {
	number textinput.Model
	date   textinput.Model
}

type profileForm struct {
	passCurr textinput.Model
	passNew  textinput.Model
}

type localUploadedData struct {
	filename string
	size     int64
}

type profileUser struct {
	email     string
	pass      string
	createdAt string
}

type arbitraryData struct {
	id         int64
	filename   string
	filetype   string
	status     int64
	objectKey  string
	osPath     string
	clientSize int64
	totalSize  int64
	createdAt  time.Time
}
