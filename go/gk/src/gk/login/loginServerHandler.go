/*
	Copyright 2012-2013 1620469 Ontario Limited.

	This program is free software: you can redistribute it and/or modify
	it under the terms of the GNU Affero General Public License as published by
	the Free Software Foundation, either version 3 of the License, or
	(at your option) any later version.

	This program is distributed in the hope that it will be useful,
	but WITHOUT ANY WARRANTY; without even the implied warranty of
	MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
	GNU Affero General Public License for more details.

	You should have received a copy of the GNU Affero General Public License
	along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

// to_do
// this program logs any errors
// but does not always forward the user to a meaningful error page
package login

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"
)

import (
	"gk/database"
	"gk/gkerr"
	"gk/gklog"
	"gk/gknet"
	"gk/gktmpl"
	"gk/sec"
)

const _methodGet = "GET"
const _methodPost = "POST"
const _loginServer = "/gk/loginServer"
const _gameServer = "/gk/gameServer"
const _tokenServer = "/gk/tokenServer"

const _actParam = "act"
const _loginParam = "login"
const _registerParam = "register"
const _forgotPasswordParam = "forgot_password"
const _userNameParam = "userName"
const _passwordParam = "password"
const _emailParam = "email"
const _tokenParam = "token"

var _loginTemplate *gktmpl.TemplateDef
var _loginTemplateName string = "login"

var _registerTemplate *gktmpl.TemplateDef
var _registerTemplateName string = "register"

var _forgotPasswordTemplate *gktmpl.TemplateDef
var _forgotPasswordTemplateName string = "forgot_password"

var _forgotPasswordEmailTemplate *gktmpl.TemplateDef
var _forgotPasswordEmailTemplateName string = "forgot_password_email"

var _resetPasswordTemplate *gktmpl.TemplateDef
var _resetPasswordTemplateName string = "reset_password"

var _errorTemplate *gktmpl.TemplateDef
var _errorTemplateName string = "error"

type loginDataDef struct {
	Title                 string
	ErrorList             []string
	UserName              string
	UserNameError         template.HTML
	PasswordError         template.HTML
	LoginWebAddressPrefix string
}

type registerDataDef struct {
	Title                 string
	ErrorList             []string
	UserName              string
	UserNameError         template.HTML
	PasswordError         template.HTML
	Email                 string
	EmailError            template.HTML
	LoginWebAddressPrefix string
}

type forgotPasswordDataDef struct {
	Title                 string
	ErrorList             []string
	UserName              string
	UserNameError         template.HTML
	LoginWebAddressPrefix string
}

type forgotPasswordEmailDataDef struct {
	UserName              string
	Token                 string
	LoginWebAddressPrefix string
}

type resetPasswordDataDef struct {
	Title                 string
	ErrorList             []string
	UserName              string
	Token                 string
	LoginWebAddressPrefix string
}

type errorDataDef struct {
	Title   string
	Message string
}

func (loginConfig *loginConfigDef) loginInit() *gkerr.GkErrDef {
	var gkErr *gkerr.GkErrDef

	//	var fileNames []string

	//	fileNames = []string{"main", "head", "error_list", "login"}
	_loginTemplate, gkErr = gktmpl.NewTemplate(loginConfig.TemplateDir, _loginTemplateName)
	if gkErr != nil {
		return gkErr
	}

	//	fileNames = []string{"main", "head", "error_list", "register"}
	_registerTemplate, gkErr = gktmpl.NewTemplate(loginConfig.TemplateDir, _registerTemplateName)
	if gkErr != nil {
		return gkErr
	}

	//	fileNames = []string{"main", "head", "error_list", "error"}
	_errorTemplate, gkErr = gktmpl.NewTemplate(loginConfig.TemplateDir, _errorTemplateName)
	if gkErr != nil {
		return gkErr
	}

	_forgotPasswordTemplate, gkErr = gktmpl.NewTemplate(loginConfig.TemplateDir, _forgotPasswordTemplateName)
	if gkErr != nil {
		return gkErr
	}

	_forgotPasswordEmailTemplate, gkErr = gktmpl.NewTemplate(loginConfig.TemplateDir, _forgotPasswordEmailTemplateName)
	if gkErr != nil {
		return gkErr
	}

	_resetPasswordTemplate, gkErr = gktmpl.NewTemplate(loginConfig.TemplateDir, _resetPasswordTemplateName)
	if gkErr != nil {
		return gkErr
	}

	return nil
}

func (loginConfig *loginConfigDef) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	if _loginTemplate == nil {
		gklog.LogError("missing call to loginInit")
	}

	path := req.URL.Path

	gklog.LogTrace(req.Method)
	gklog.LogTrace(path)

	if req.Method == _methodGet || req.Method == _methodPost {
		if gknet.RequestMatches(path, _loginServer) {
			handleLogin(loginConfig, res, req)
		} else {
			http.NotFound(res, req)
		}
	} else {
		http.NotFound(res, req)
	}

}

func handleLogin(loginConfig *loginConfigDef, res http.ResponseWriter, req *http.Request) {
	var act string
	var userName string
	var password string
	var email string
	var token string

	req.ParseForm()

	act = req.Form.Get(_actParam)
	userName = req.Form.Get(_userNameParam)
	password = req.Form.Get(_passwordParam)
	email = req.Form.Get(_emailParam)
	token = req.Form.Get(_tokenParam)

	// for security sleep on password attempt
	time.Sleep(sec.GetSleepDurationPasswordAttempt())

	gklog.LogTrace("act: " + act)

	switch act {
	case "":
		var login string

		login = req.Form.Get(_loginParam)

		if login != "" {
			handleLoginLogin(loginConfig, res, req, userName, password)
			return
		}

		handleLoginInitial(loginConfig, res, req)
		return
	case "login":
		var register string
		var forgotPassword string

		register = req.Form.Get(_registerParam)
		forgotPassword = req.Form.Get(_forgotPasswordParam)

		if register != "" {
			handleLoginRegisterInitial(loginConfig, res, req, userName)
			return
		}

		if forgotPassword != "" {
			handleLoginForgotPasswordInitial(loginConfig, res, req)
			return
		}

		if userName == "" {
			handleLoginInitial(loginConfig, res, req)
			return
		}
		handleLoginLogin(loginConfig, res, req, userName, password)
		return
	case "register":
		handleLoginRegister(loginConfig, res, req, userName, password, email)
	case "forgot_password":
		handleLoginForgotPassword(loginConfig, res, req, userName)
	case "reset_password":
		handleLoginResetPassword(loginConfig, res, req, token, userName, password)
	default:
		gklog.LogError("unknown act")
		redirectToError("unknown act", res, req)
		return
	}
}

func handleLoginInitial(loginConfig *loginConfigDef, res http.ResponseWriter, req *http.Request) {
	var loginData loginDataDef
	var gkErr *gkerr.GkErrDef

	loginData.Title = "login"
	loginData.LoginWebAddressPrefix = loginConfig.LoginWebAddressPrefix

	gkErr = _loginTemplate.Build(loginData)
	if gkErr != nil {
		gklog.LogGkErr("_loginTemplate.Build", gkErr)
		redirectToError("_loginTemplate.Build", res, req)
		return
	}

	gkErr = _loginTemplate.Send(res, req)
	if gkErr != nil {
		gklog.LogGkErr("_loginTemplate.Send", gkErr)
		return
	}
}

func handleLoginLogin(loginConfig *loginConfigDef, res http.ResponseWriter, req *http.Request, userName string, password string) {
	var loginData loginDataDef
	var gkErr *gkerr.GkErrDef
	var gotError bool

	loginData.Title = "login"
	loginData.UserName = userName
	loginData.LoginWebAddressPrefix = loginConfig.LoginWebAddressPrefix

	if loginData.UserName == "" {
		loginData.ErrorList = append(loginData.ErrorList, "invalid user name")
		loginData.UserNameError = genErrorMarker()
		gotError = true
	}

	if password == "" {
		loginData.ErrorList = append(loginData.ErrorList, "invalid password")
		loginData.PasswordError = genErrorMarker()
		gotError = true
	}

	var passwordHashFromUser []byte

	var dbUser *database.DbUserDef
	var gkDbCon *database.GkDbConDef

	if !gotError {

		gkDbCon, gkErr = database.NewGkDbCon(loginConfig.DatabaseUserName, loginConfig.DatabasePassword, loginConfig.DatabaseHost, loginConfig.DatabasePort, loginConfig.DatabaseDatabase)
		if gkErr != nil {
			gklog.LogGkErr("database.NewGkDbCon", gkErr)
			redirectToError("database.NewGkDbCon", res, req)
			return
		}

		defer gkDbCon.Close()

		dbUser, gkErr = gkDbCon.GetUser(loginData.UserName)

		if gkErr != nil {
			if gkErr.GetErrorId() == database.ERROR_ID_NO_ROWS_FOUND {
				var passwordSalt string

				password = "one two three"
				passwordSalt = "abc123QWE."
				// make it take the same amount of time
				// between no user and invalid password
				passwordHashFromUser = sec.GenPasswordHashSlow([]byte(password), []byte(passwordSalt))
				loginData.ErrorList = append(loginData.ErrorList, "invalid username/password")
				loginData.UserNameError = genErrorMarker()
				loginData.PasswordError = genErrorMarker()
				gotError = true
			} else {
				gklog.LogGkErr("gkDbCon.GetPasswordHashAndSalt", gkErr)
				redirectToError("gkDbCon.GetPasswordhashAndSalt", res, req)
				return
			}
		}
	}

	if !gotError {
		passwordHashFromUser = sec.GenPasswordHashSlow([]byte(password), []byte(dbUser.PasswordSalt))

		gklog.LogTrace(fmt.Sprintf("dbUser: %v fromUser: %s", dbUser, passwordHashFromUser))
		if dbUser.PasswordHash != string(passwordHashFromUser) {
			loginData.ErrorList = append(loginData.ErrorList, "invalid username/password")
			loginData.UserNameError = genErrorMarker()
			loginData.PasswordError = genErrorMarker()
			gotError = true
		}
	}

	if gotError {
		// for security, to slow down an attack that is guessing passwords,
		// sleep between 100 and 190 milliseconds
		time.Sleep(sec.GetSleepDurationPasswordInvalid())

		gkErr = _loginTemplate.Build(loginData)
		if gkErr != nil {
			gklog.LogGkErr("_loginTemplate.Build", gkErr)
			redirectToError("_loginTemplate.Build", res, req)
			return
		}

		gkErr = _loginTemplate.Send(res, req)
		if gkErr != nil {
			gklog.LogGkErr("_loginTemplate.Send", gkErr)
			return
		}
	} else {
		gkErr = gkDbCon.UpdateUserLoginDate(dbUser.UserName)
		if gkErr != nil {
			// this error is going to be logged
			// but the user is not going to be redirected to an error
			// because they are going to be redirected to the game server
			// and it is not critical that their login date be updated.
			gklog.LogGkErr("_loginTemplate.Send", gkErr)
		}
		var gameRedirect string
		gameRedirect, gkErr = getGameRedirect(loginConfig, loginData.UserName)
		if gkErr != nil {
			gklog.LogGkErr("getGameRedirect", gkErr)
			return
		}
		http.Redirect(res, req, gameRedirect, http.StatusFound)
	}
}

func getGameRedirect(loginConfig *loginConfigDef, userName string) (string, *gkerr.GkErrDef) {
	var newSessionToken string
	var gkErr *gkerr.GkErrDef

	newSessionToken, gkErr = getNewSessionToken(loginConfig, userName)

	if gkErr != nil {
		return "", gkErr
	}

	return loginConfig.GameWebAddressPrefix + _gameServer + "?token=" + newSessionToken, nil
}

func getNewSessionToken(loginConfig *loginConfigDef, userName string) (string, *gkerr.GkErrDef) {
	var res *http.Response
	var err error
	var gkErr *gkerr.GkErrDef

	res, err = http.Get(loginConfig.GameTokenAddressPrefix + _tokenServer + "?userName=" + url.QueryEscape(userName))
	if err != nil {
		gkErr = gkerr.GenGkErr("token http.Get", err, ERROR_ID_TOKEN_SERVER_GET)
		return "", gkErr
	}

	defer res.Body.Close()

	var token []byte
	token, err = ioutil.ReadAll(res.Body)
	if err != nil {
		gkErr = gkerr.GenGkErr("token ioutil.ReadAll", err, ERROR_ID_TOKEN_SERVER_READ)
		return "", gkErr
	}

	return string(token), nil
}

func handleLoginRegisterInitial(loginConfig *loginConfigDef, res http.ResponseWriter, req *http.Request, userName string) {
	var registerData registerDataDef
	var gkErr *gkerr.GkErrDef

	registerData.Title = "register"
	registerData.LoginWebAddressPrefix = loginConfig.LoginWebAddressPrefix
	registerData.UserName = userName

	gkErr = _registerTemplate.Build(registerData)
	if gkErr != nil {
		gklog.LogGkErr("_registerTemplate.Build", gkErr)
		redirectToError("_registerTemplate.Build", res, req)
		return
	}

	gkErr = _registerTemplate.Send(res, req)
	if gkErr != nil {
		gklog.LogGkErr("_registerTemplate.send", gkErr)
	}
}

func handleLoginRegister(loginConfig *loginConfigDef, res http.ResponseWriter, req *http.Request, userName string, password string, email string) {
	var registerData registerDataDef
	var gkErr *gkerr.GkErrDef
	var err error

	registerData.Title = "register"
	registerData.LoginWebAddressPrefix = loginConfig.LoginWebAddressPrefix
	registerData.UserName = userName
	registerData.Email = email
	registerData.ErrorList = make([]string, 0, 0)

	var gotError bool

	if !isNewUserNameValid(userName) {
		registerData.ErrorList = append(registerData.ErrorList, "invalid user name")
		registerData.UserNameError = genErrorMarker()
		gotError = true
	}
	if !isPasswordValid(password) {
		registerData.ErrorList = append(registerData.ErrorList, "invalid password")
		registerData.PasswordError = genErrorMarker()
		gotError = true
	}
	if !isEmailValid(email) {
		registerData.ErrorList = append(registerData.ErrorList, "invalid email")
		registerData.EmailError = genErrorMarker()
		gotError = true
	}

	if !gotError {
		var gkDbCon *database.GkDbConDef

		gkDbCon, gkErr = database.NewGkDbCon(loginConfig.DatabaseUserName, loginConfig.DatabasePassword, loginConfig.DatabaseHost, loginConfig.DatabasePort, loginConfig.DatabaseDatabase)
		if gkErr != nil {
			gklog.LogGkErr("database.NewGkDbCon", gkErr)
			redirectToError("database.NewGkDbCon", res, req)
			return
		}

		defer gkDbCon.Close()

		var passwordHash, passwordSalt []byte

		passwordSalt, err = sec.GenSalt()
		if err != nil {
			gkErr = gkerr.GenGkErr("sec.GenSalt", err, ERROR_ID_GEN_SALT)
			gklog.LogGkErr("sec.GenSalt", gkErr)
			redirectToError("sec.GenSalt", res, req)
		}

		passwordHash = sec.GenPasswordHashSlow([]byte(password), passwordSalt)

		gkErr = gkDbCon.AddNewUser(
			registerData.UserName,
			string(passwordHash),
			string(passwordSalt),
			email)

		if gkErr != nil {
			if gkErr.GetErrorId() == database.ERROR_ID_UNIQUE_VIOLATION {
				registerData.ErrorList = append(registerData.ErrorList, "user name already in use")
				registerData.UserNameError = genErrorMarker()
				gotError = true
			} else {
				gklog.LogGkErr("gbDbCon.AddNewUser", gkErr)
				redirectToError("gbDbCon.AddNewUser", res, req)
				return
			}
		}
	}

	if gotError {
		gkErr = _registerTemplate.Build(registerData)
		if gkErr != nil {
			gklog.LogGkErr("_registerTemplate.Build", gkErr)
			redirectToError("_registerTemplate.Build", res, req)
			return
		}

		gkErr = _registerTemplate.Send(res, req)
		if gkErr != nil {
			gklog.LogGkErr("_registerTemplate.send", gkErr)
		}
	} else {
		http.Redirect(res, req, _loginServer, http.StatusFound)
		//		var gameRedirect string
		//		gameRedirect = getGameRedirect(loginConfig, loginData.UserName)
		//		http.Redirect(res, req, gameRedirect, http.StatusFound)
	}
}

func handleLoginForgotPasswordInitial(loginConfig *loginConfigDef, res http.ResponseWriter, req *http.Request) {
	var forgotPasswordData forgotPasswordDataDef
	var gkErr *gkerr.GkErrDef

	forgotPasswordData.Title = "forgotPassword"
	forgotPasswordData.LoginWebAddressPrefix = loginConfig.LoginWebAddressPrefix

	gkErr = _forgotPasswordTemplate.Build(forgotPasswordData)
	if gkErr != nil {
		gklog.LogGkErr("_forgotPasswordTemplate.Build", gkErr)
		redirectToError("_forgotPasswordTemplate.Build", res, req)
		return
	}

	gkErr = _forgotPasswordTemplate.Send(res, req)
	if gkErr != nil {
		gklog.LogGkErr("_forgotPasswordTemplate.send", gkErr)
	}
}

func handleLoginForgotPassword(loginConfig *loginConfigDef, res http.ResponseWriter, req *http.Request, userName string) {
	var forgotPasswordData forgotPasswordDataDef
	var gkErr *gkerr.GkErrDef

	forgotPasswordData.Title = "forgotPassword"
	forgotPasswordData.LoginWebAddressPrefix = loginConfig.LoginWebAddressPrefix
	forgotPasswordData.UserName = userName
	forgotPasswordData.ErrorList = make([]string, 0, 0)

	var gotError bool

	if userName == "" {
		forgotPasswordData.ErrorList = append(forgotPasswordData.ErrorList, "user name cannot be blank")
		forgotPasswordData.UserNameError = genErrorMarker()
		gotError = true
	}

	var dbUser *database.DbUserDef

	if !gotError {
		var gkDbCon *database.GkDbConDef

		gkDbCon, gkErr = database.NewGkDbCon(loginConfig.DatabaseUserName, loginConfig.DatabasePassword, loginConfig.DatabaseHost, loginConfig.DatabasePort, loginConfig.DatabaseDatabase)
		if gkErr != nil {
			gklog.LogGkErr("database.NewGkDbCon", gkErr)
			redirectToError("database.NewGkDbCon", res, req)
			return
		}

		defer gkDbCon.Close()

		dbUser, gkErr = gkDbCon.GetUser(
			forgotPasswordData.UserName)

		if gkErr != nil {
			if gkErr.GetErrorId() == database.ERROR_ID_NO_ROWS_FOUND {
				forgotPasswordData.ErrorList = append(forgotPasswordData.ErrorList, "no such user")
				forgotPasswordData.UserNameError = genErrorMarker()
				gotError = true
			} else {
				gklog.LogGkErr("gbDbCon.GetUser", gkErr)
				redirectToError("gbDbCon.GetUser", res, req)
				return
			}
		}
	}

	var err error

	if !gotError {
		// create temporary forgot password token

		//var token []byte
		var forgotPasswordEmailData forgotPasswordEmailDataDef

		forgotPasswordEmailData.LoginWebAddressPrefix = loginConfig.LoginWebAddressPrefix
		forgotPasswordEmailData.UserName = userName

		var token []byte

		token, err = sec.GenForgotPasswordToken()
		if err != nil {
			gkErr = gkerr.GenGkErr("GenForgotPasswordToken", err, ERROR_ID_GEN_TOKEN)
			gklog.LogGkErr("GenForgotPasswordToken", gkErr)
			redirectToError("GenForgotPasswordToken", res, req)
			return
		}

		forgotPasswordEmailData.Token = string(token)

		gkErr = _forgotPasswordEmailTemplate.Build(forgotPasswordEmailData)
		if gkErr != nil {
			gklog.LogGkErr("_forgotPasswordEmailTemplate.Build", gkErr)
			redirectToError("_forgotPasswordEmailTemplate.Build", res, req)
			return
		}

		var message []byte

		message, gkErr = _forgotPasswordEmailTemplate.GetBytes()
		if gkErr != nil {
			gklog.LogGkErr("_forgotPasswordEmailTemplate.GetBytes", gkErr)
			redirectToError("_forgotPasswordEmailTemplate.GetBytes", res, req)
			return
		}

		toArray := make([]string, 1, 1)
		toArray[0] = dbUser.Email
		var sendId string

		AddNewToken(string(token), userName)

		sendId, gkErr = gknet.SendEmail(loginConfig.EmailServer, loginConfig.ServerFromEmail, toArray, "gourdian knot forgotten password", message)

		if gkErr != nil {
			gklog.LogGkErr("gknet.SendEmail", gkErr)
		} else {
			gklog.LogTrace("forgot email sent to: " + toArray[0] + " sendId: [" + sendId + "]")
		}
	}

	if gotError {
		gkErr = _forgotPasswordTemplate.Build(forgotPasswordData)
		if gkErr != nil {
			gklog.LogGkErr("_forgotPasswordTemplate.Build", gkErr)
			redirectToError("_forgotPasswordTemplate.Build", res, req)
			return
		}

		gkErr = _forgotPasswordTemplate.Send(res, req)
		if gkErr != nil {
			gklog.LogGkErr("_forgotPasswordTemplate.send", gkErr)
		}
	} else {
		http.Redirect(res, req, _loginServer, http.StatusFound)
	}
}

func handleLoginResetPassword(loginConfig *loginConfigDef, res http.ResponseWriter, req *http.Request, token string, userName string, password string) {
	var resetPasswordData resetPasswordDataDef
	var gkErr *gkerr.GkErrDef

	resetPasswordData.Title = "resetPassword"
	resetPasswordData.LoginWebAddressPrefix = loginConfig.LoginWebAddressPrefix
	resetPasswordData.Token = token
	resetPasswordData.UserName = userName

	if !CheckToken(token, userName) {
		redirectToError("token expired", res, req)
		return
	}

	gklog.LogTrace("reset password: " + password)
	if password == "" {
		gklog.LogTrace("password blank")
		gkErr = _resetPasswordTemplate.Build(resetPasswordData)
		if gkErr != nil {
			gklog.LogGkErr("_resetPasswordTemplate.Build", gkErr)
			redirectToError("_resetPasswordTemplate.Build", res, req)
			return
		}

		gkErr = _resetPasswordTemplate.Send(res, req)
		if gkErr != nil {
			gklog.LogGkErr("_resetPasswordTemplate.send", gkErr)
		}
		return
	}

	var gkDbCon *database.GkDbConDef

	gkDbCon, gkErr = database.NewGkDbCon(loginConfig.DatabaseUserName, loginConfig.DatabasePassword, loginConfig.DatabaseHost, loginConfig.DatabasePort, loginConfig.DatabaseDatabase)
	if gkErr != nil {
		gklog.LogGkErr("database.NewGkDbCon", gkErr)
		redirectToError("database.NewGkDbCon", res, req)
		return
	}

	defer gkDbCon.Close()

	var passwordHash, passwordSalt []byte
	var err error

	passwordSalt, err = sec.GenSalt()
	if err != nil {
		gkErr = gkerr.GenGkErr("sec.GenSalt", err, ERROR_ID_GEN_SALT)
		gklog.LogGkErr("sec.GenSalt", gkErr)
		redirectToError("sec.GenSalt", res, req)
	}

	passwordHash = sec.GenPasswordHashSlow([]byte(password), passwordSalt)

	gklog.LogTrace("change password")
	gkDbCon.ChangePassword(userName, string(passwordHash), string(passwordSalt))
	if gkErr != nil {
		gklog.LogGkErr("gkDbCon.ChangePassword", gkErr)
		redirectToError("gbDbCon.ChangePassword", res, req)
		return
	}

	gklog.LogTrace("redirect to login")
	http.Redirect(res, req, loginConfig.LoginWebAddressPrefix+_loginServer, http.StatusFound)
}

func genErrorMarker() template.HTML {
	return template.HTML("<span class=\"errorMarker\">*</span>")
}

func redirectToError(message string, res http.ResponseWriter, req *http.Request) {
	var errorData errorDataDef
	var gkErr *gkerr.GkErrDef

	errorData.Title = "Error"
	errorData.Message = message

	gkErr = _errorTemplate.Build(errorData)
	if gkErr != nil {
		gklog.LogGkErr("_errorTemplate.Build", gkErr)
	}

	gkErr = _errorTemplate.Send(res, req)
	if gkErr != nil {
		gklog.LogGkErr("_errorTemplate.Send", gkErr)
	}
}
