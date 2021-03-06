package auth

import (
	"bytes"
	"html/template"
)

type data struct {
	URL template.URL
}

func renderEmail(tpl string, data interface{}) (string, error) {
	t, err := template.New("").Parse(tpl)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	if err := t.Execute(&buf, data); err != nil {
		return "", err
	}

	return buf.String(), nil
}

const loginTpl = `
<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="IE=7" />
    <title>Get Started</title>
    <style>
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: "Microsoft YaHei", "微软雅黑", STXihei;
            background: white;
            color: black;
            line-height: 1.8em;
        }

        a {
            text-decoration: none;
        }

        #container {
            border: black solid 1px;
            max-width: 30%;
            margin: 30px auto;
            padding: 30px;
            background-color: white;
        }

        .form-wrap {
            background: white;
            padding: 30px;
        }

        .form-wrap h1,
        .form-wrap p {
            margin-top: 30px;
            text-align: left;
        }

        .form-wrap .form-group {
            margin-top: 10px;
            text-align: center;
        }

        .form-wrap .form-group a {
            display: block;
            width: 200px;
            padding: 10px;
            margin-top: 10px;
            border: black 2px solid;
            border-radius: 10px;
        }

        .form-wrap button {
            /* display: inline; */
            width: 50%;
            text-align: center;
            padding: 10px;
            margin: 20px auto;
            background: white;
            cursor: pointer;
            border-radius: 10px;
        }

        .form-wrap button:hover {
            background: darkgray;
        }

        .form-wrap .bottom-text {
            font-size: 3px;
            text-align: left;
            /* margin-top: 20px; */
        }

        .form-wrap .bottom-text a {
            font-size: 5px;
            text-align: left;
            /* margin-top: 20px; */
        }

        .small-footer {
            text-align: center;
            margin-top: 5px;
            font-size: 1px;
            color: green;
        }

        .big-logo-head {
            font-size: 50px;
            margin-bottom: 50px;
        }
    </style>
</head>

<body>
    <div id="container">
        <div class="form-wrap">
            <h1 class="big-logo-head">example</h1>
            <br />
            <h2>You're almost there</h2>
            <p>
                Click the link below to sign in to your Medium account.
            </p>
            <p>This link will expire in 2 hours and can only be used once.</p>
            <form>
                <div class="form-group">
                    <a href="{{.URL}}">Sign in to Example</a>
                </div>
                <p class="bottom-text">
                    If the button above doesn’t work, paste this link into your web
                    browser:
                    <a href="{{.URL}}">{{.URL}}</a>
                </p>
            </form>
            <p class="bottom-text">
                If you did not make this request, you can safely ignore this email.
            </p>
            <hr />
            <p class="bottom-text">
                Sent by Example
                <br />
                <a class="small-footer" href="#">· Careers</a>
                <a class="small-footer" href="#">· Help center</a>
                <br />
                <a class="small-footer" href="#">· Privacy policy</a>
                <a class="small-footer" href="#">· Terms of service</a>
            </p>
        </div>
    </div>
</body>

</html>
`

const registerTpl = `
<!DOCTYPE html>
<html lang="en">

<head>
    <meta charset="UTF-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0" />
    <meta http-equiv="X-UA-Compatible" content="IE=7" />
    <title>Get Started</title>
    <style>
        * {
            box-sizing: border-box;
            margin: 0;
            padding: 0;
        }

        body {
            font-family: "Microsoft YaHei", "微软雅黑", STXihei;
            background: white;
            color: black;
            line-height: 1.8em;
        }

        a {
            text-decoration: none;
        }

        #container {
            border: black solid 1px;
            max-width: 30%;
            margin: 30px auto;
            padding: 30px;
            background-color: white;
        }

        .form-wrap {
            background: white;
            padding: 30px;
        }

        .form-wrap h1,
        .form-wrap p {
            margin-top: 30px;
            text-align: left;
        }

        .form-wrap .form-group {
            margin-top: 10px;
            text-align: center;
        }

        .form-wrap .form-group a {
            display: block;
            width: 200px;
            padding: 10px;
            margin-top: 10px;
            border: black 2px solid;
            border-radius: 10px;
        }

        .form-wrap button {
            /* display: inline; */
            width: 50%;
            text-align: center;
            padding: 10px;
            margin: 20px auto;
            background: white;
            cursor: pointer;
            border-radius: 10px;
        }

        .form-wrap button:hover {
            background: darkgray;
        }

        .form-wrap .bottom-text {
            font-size: 3px;
            text-align: left;
            /* margin-top: 20px; */
        }

        .form-wrap .bottom-text a {
            font-size: 5px;
            text-align: left;
            /* margin-top: 20px; */
        }

        .small-footer {
            text-align: center;
            margin-top: 5px;
            font-size: 1px;
            color: green;
        }

        .big-logo-head {
            font-size: 50px;
            margin-bottom: 50px;
        }
    </style>
</head>

<body>
    <div id="container">
        <div class="form-wrap">
            <h1 class="big-logo-head">example</h1>
            <br />
            <h2>You're almost there</h2>
            <p>
                Click the link below to confirm your email and finish creating your
                Example account.
            </p>
            <p>This link will expire in 2 hours and can only be used once.</p>
            <form>
                <div class="form-group">
                    <a href="{{.URL}}">Create your account</a>
                </div>
                <p class="bottom-text">
                    If the button above doesn’t work, paste this link into your web
                    browser:
                    <a href="{{.URL}}">{{.URL}}</a>
                </p>
            </form>
            <p class="bottom-text">
                If you did not make this request, you can safely ignore this email.
            </p>
            <hr />
            <p class="bottom-text">
                Sent by Example
                <br />
                <a class="small-footer" href="#">· Careers</a>
                <a class="small-footer" href="#">· Help center</a>
                <br />
                <a class="small-footer" href="#">· Privacy policy</a>
                <a class="small-footer" href="#">· Terms of service</a>
            </p>
        </div>
    </div>
</body>

</html>
`
