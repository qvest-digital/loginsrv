package login

import (
	"html/template"
	"io"
)

const loginForm = `<!DOCTYPE html>
<html>
  <head>
    <link uic-remove rel="stylesheet" href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.5/css/bootstrap.min.css">
    <style>
      .vertical-offset-100{
        padding-top:100px;
      }
    </style>
  </head>
  <body>
    <uic-fragment name="content">
<div class="container">
    <div class="row vertical-offset-100">
    	<div class="col-md-4 col-md-offset-4">
    		<div class="panel panel-default">
			  	<div class="panel-heading">
  			    	  <h3 class="panel-title">Please sign in</h3>
                                  {{ if .error}}Internal Error. Please try again later{{end}} 
                                  {{ if .failure}}Wrong credentials{{end}} 
			 	</div>
			  	<div class="panel-body">
			    	<form accept-charset="UTF-8" role="form" method="POST" action="{{.path}}">
                                 <fieldset>
			    	  	<div class="form-group">
			    		    <input class="form-control" placeholder="Username" name="username" value="{{.username}}" type="text">
			    		</div>
			    		<div class="form-group">
			    	            <input class="form-control" placeholder="Password" name="password" type="password" value="">
			    		</div>
			    		<input class="btn btn-lg btn-success btn-block" type="submit" value="Login">
			    	</fieldset>
			      	</form>
			    </div>
			</div>
		</div>
	</div>
</div>
    </uic-fragment>
  </body>
</html>
`

func writeLoginForm(w io.Writer, params map[string]interface{}) {
	t := template.Must(template.New("loginForm").Parse(loginForm))
	t.Execute(w, params)
}
