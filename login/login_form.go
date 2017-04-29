package login

import (
	"bytes"
	"github.com/tarent/lib-compose/logging"
	"html/template"
	"net/http"
)

const loginForm = `<!DOCTYPE html>
<html>
  <head>
    <link uic-remove rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css">
    <link uic-remove rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/bootstrap-social/5.1.1/bootstrap-social.min.css">
    <link uic-remove rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/4.7.0/css/font-awesome.css">
    <style>
     .vertical-offset-100{
       padding-top:100px;
     }
     .login-or-container {
       text-align: center;
       margin: 0;
       margin-bottom: 10px;
       clear: both;
       color: #6a737c;
       font-variant: small-caps;
     }
     .login-or-hr {
       margin-bottom: 0;
       position: relative;
       top: 28px;
       height: 0;
       border: 0;
       border-top: 1px solid #e4e6e8;
     }
     .login-or {
       display: inline-block;
       position: relative;
       padding: 10px;
       background-color: #FFF;
     }
    </style>
  </head>
  <body>
    <uic-fragment name="content">
      <div class="container">
        <div class="row vertical-offset-100">
    	  <div class="col-md-4 col-md-offset-4">

            {{ if .Error}}
              <div class="alert alert-danger" role="alert">
                <strong>Internal Error. </strong> Please try again later.
              </div>
            {{end}}

            {{ if .Authenticated}}
              <h3>Welcome {{.UserInfo.Username}}</h3>
              <a href="login?logout=true">Logout</a>
            {{else}}
              <a class="btn btn-block btn-lg btn-social btn-github" href="login/github">
                <span class="fa fa-github"></span> Sign in with Github
              </a>
              <div class="login-or-container">
                <hr class="login-or-hr">
                <div class="login-or lead">or</div>
              </div>
              <div class="panel panel-default">
  	        <div class="panel-heading">
  
  		  <div class="panel-title">
  		    <h4>Sign in</h4>
                    {{ if .Failure}}<div class="alert alert-warning" role="alert">Invalid credentials</div>{{end}} 
		  </div>
	        </div>
	        <div class="panel-body">
		  <form accept-charset="UTF-8" role="form" method="POST" action="{{.Path}}">
                    <fieldset>
		      <div class="form-group">
		        <input class="form-control" placeholder="Username" name="username" value="{{.UserInfo.Username}}" type="text">
		      </div>
		      <div class="form-group">
		        <input class="form-control" placeholder="Password" name="password" type="password" value="">
		      </div>
		      <input class="btn btn-lg btn-success btn-block" type="submit" value="Login">
		    </fieldset>
		  </form>
	        </div>
	      </div>
            {{end}}
	  </div>
	</div>
      </div>
    </uic-fragment>
  </body>
</html>`

type loginFormData struct {
	Path          string
	Error         bool
	Failure       bool
	Config        *Config
	Authenticated bool
	UserInfo      UserInfo
}

func writeLoginForm(w http.ResponseWriter, params loginFormData) {
	t := template.Must(template.New("loginForm").Parse(loginForm))
	b := bytes.NewBuffer(nil)
	err := t.Execute(b, params)
	if err != nil {
		logging.Logger.WithError(err).Error()
		w.WriteHeader(500)
		w.Write([]byte(`Internal Server Error`))
		return
	}

	w.Header().Set("Content-Type", contentTypeHtml)
	if params.Error {
		w.WriteHeader(500)
	}

	w.Write(b.Bytes())
}
