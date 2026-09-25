# Golang Testing JSON and HTML testing

Look, I hate the way testing looks. The aesthetic of it all makes it difficult to work out what is actually wrong and when.
I went hardcore TDD for a while, unit test everything, write tets before code ...
But just found that anytime my cod had to change to adapt, all my tests failed and had to be changed as well.

Now, when I write a project I go for the public interface tests as quickly as possible. This tests a lot of how I want users to interact,
and that way when I update code it is because it impacts users. 

Anyway, I made a DSL that I like for testing JSON and HTML endpoints that (certain) people might like.

# JSON DSL

When I am writing tests for JSON I am mostly testing that a series of JSON calls do a thing. 
I don't care how they do it, but the contract is that it is accomplised.

```
Store this model
Get the model to make sure it is stored
Update that model
Get that mode to make sure it updated
Search for models with params that wont return the model
Search for models with params that will return the model
```

All this kind of crap. I found it was leading to a lot of go code that was just being repeated, like setting up headers, and setting up requests and parsing output...

So I made this thing based on the `.http` format from <TODO>

```
NAME store_model
POST {{ site }}/model
H Authorizatio Bearer: {{ jwt }}
B {
    "key": "value"
}
ASSERT_STATUS 201
```

It is relying on the golang templates

Now the next query will need some information, like the id from the previous model

```
NAME get_model
GET {{ site }}/model/{{ dig ".site_model.id" }}
H Authorizatio Bearer: {{ jwt }}
ASSERT_EQUAL @key "value"
```

Wait, how does auth work I see that `jwt` thing and I want to know?

```
NAME jwtauth
GET http://{{ site }}/api/v2/authorization/request/token={{ token }}
ASSERT_STATUS 200
SET jwt {{ .jwtauth.token }}
```

Here you can see the `SET` which is just a convience for naming things.
Each query builds a `ctx` of type `map[string]any` to the next method, and you can start the test by injecting that.

Lets see how that looks.:

```
func RunTheRequestFileAsATest(t *testing.T, path string, router http.Handler, defaultJwt string) {
	t.Run(removeNightscout, func(t *testing.T) {
		requestFileBytes, err := embeddedFiles.ReadFile(path)
		if err != nil {
			t.Fatalf("failed to read request file: %v", err)
		}
		fmt.Println()
		fmt.Println("===============================================")
		fmt.Printf("Running request file: %s\n", path)
		fmt.Println("===============================================")

		requestContext := map[string]any{
			"default_site":  siteInformation.SiteDB.Subdomain + ".ns.localhost",
			"default_token": siteInformation.PwdNSToken,
			"default_jwt":   defaultJwt,
		}
		requestFile, err := jsonapi.ParseRequestFile(string(requestFileBytes), requestContext, path)
		if err != nil {
			t.Fatalf("failed to parse request file: %v", err)
		}

		for _, request := range requestFile.Requests {

			if err != nil {
				t.Fatalf("failed to build real request: %v", err)
			}

			newContext, err := request.ExecuteRequest(t, router, requestContext)
			if err != nil {
				t.Fatalf("failed to execute request: %v, method: %s, url: %s", err, request.Method, request.URL)
            }
			requestContext = newContext
		}
	})
}
```

What I have done is have a folder filled with these tests and just walk the path and run them all as tests.

I add a bunch of `fmt.Println` in this loop to seperate out tests and print responses so that I can see what is going on.


# HTML 

So I made this JSON API testing framework and needed something similar for HTML that just went through a series of steps to check everything is working.
It doesnt quite work the same with HTML things are less transactional, so it looks a bit different.

```
GOTO {{ server_url}}/
ASSERT_STATUS           200
ASSERT_VISIBLE          "Welcome"

CLICK #login-button
ASSERT_NEW_PAGE
ASSERT_PATH /login

// PUT IN THE WRONG password
ASSERT_VISIBLE #login-name
ASSERT_VISIBLE #login-password
INPUT #login-name admin
INPUT #login-password wrongpassword

CLICK #login-button
ASSERT_STATUS 401
ASSERT_VISIBLE "Incorrect username or password"

// PUT IN THE WRONG TOTP 
ASSERT_VISIBLE #login-name
ASSERT_VISIBLE #login-password
INPUT #login-name admin
INPUT #login-password password


CLICK #login-button
ASSERT_STATUS 200
ASSERT_VISIBLE #login-otp-1
ASSERT_PATH /login
```