---
title: "Error output"
description: "The exact format CTXError.Error() prints, how every wrap layer adds its own file, line and function, and what Go names closures in the trace."
nav_order: 3
---

# Error output

When shit hits the fan, you get detailed context. `Error()` prints the message, then the wrapped error's own text, then the location in brackets:

```
<message>: <wrapped error> [<file>:<line> in <function>]
```

An error from `New()` has nothing wrapped, so it is just the message and the location:

```
<message> [<file>:<line> in <function>]
```

The file is the path the Go compiler recorded for the source file, absolute unless you build with `-trimpath`, and the function is the fully qualified name from the runtime, such as `main.connectToDatabase` or `github.com/you/app/store.(*Repo).Get`. A single wrap looks like this:

```
failed to connect to database at localhost:5432: dial tcp [::1]:5432: connect: connection refused [/path/to/your/file.go:42 in main.connectToDatabase]
```

When the [error map](error-mapping.md) translated the wrapped error, the middle part is the translation, not the original message. That keeps driver text out of your logs.

## Error chaining

When you wrap ctxerrors in a chain, each layer shows its context:

```go
func readConfig() error {
    return ctxerrors.New("config file missing")
}

func initDatabase() error {
    if err := readConfig(); err != nil {
        return ctxerrors.Wrap(err, "failed to read config")
    }
    return nil
}

func startServer() error {
    if err := initDatabase(); err != nil {
        return ctxerrors.Wrap(err, "database initialization failed")
    }
    return nil
}
```

Messages read outermost first, and locations read innermost first, because every layer appends its bracket after the text of the layer it wraps. It prints one line; it is broken up here so you can read it:

```
database initialization failed: failed to read config: config file missing
[/path/to/server.go:13 in main.readConfig]
[/path/to/server.go:18 in main.initDatabase]
[/path/to/server.go:25 in main.startServer]
```

## Stupid inline chaining

Or if you're a masochist and like one-liners:

```go
func clusterfuck() error {
    return ctxerrors.Wrap(
        ctxerrors.Wrap(
            ctxerrors.New("original fuckup"),
            "second layer of shit"),
        "final layer of despair")
}
```

Each call records the line it starts on, so the three locations differ even inside one statement:

```
final layer of despair: second layer of shit: original fuckup
[/path/to/file.go:42 in main.clusterfuck]
[/path/to/file.go:41 in main.clusterfuck]
[/path/to/file.go:40 in main.clusterfuck]
```

## Annoyingly complex tangled bullshit

Because sometimes you write code like a fucking maniac. Closures get compiler-generated names, `func1`, `func2` and so on, in the order they appear in the enclosing function:

```go
func processUserShitWithStupidNesting(userID int) error {
    validateUser := func(id int) error {
        if id <= 0 {
            return ctxerrors.New("invalid user ID: must be greater than zero")
        }
        return nil
    }

    fetchUserData := func(id int) error {
        if rand.Intn(3) == 0 {
            return ctxerrors.Wrapf(
                errors.New("connection timeout"),
                "failed to fetch user data for user ID %d", id)
        }
        return nil
    }

    processPermissions := func(id int) error {
        checkAdminRights := func() error {
            if rand.Intn(2) == 0 {
                return ctxerrors.New("user lacks admin privileges")
            }
            return nil
        }

        if err := checkAdminRights(); err != nil {
            return ctxerrors.Wrapf(err, "permission check failed for user %d", id)
        }
        return nil
    }

    // Chain all this shit together
    if err := validateUser(userID); err != nil {
        return ctxerrors.Wrap(err, "user validation step failed")
    }

    if err := fetchUserData(userID); err != nil {
        return ctxerrors.Wrap(err, "data fetching step failed")
    }

    if err := processPermissions(userID); err != nil {
        return ctxerrors.Wrap(err, "permission processing step failed")
    }

    return nil
}
```

When `fetchUserData` fails, the plain `errors.New` in the middle has no location, and the two ctxerrors layers each bring their own:

```
data fetching step failed: failed to fetch user data for user ID 42: connection timeout
[/path/to/main.go:20 in main.processUserShitWithStupidNesting.func2]
[/path/to/main.go:47 in main.processUserShitWithStupidNesting]
```

## Ridiculously stupid chain of doom

For when you really want to piss off your future self:

```go
func performStupidlyComplexOperation() error {
    // Because apparently we hate ourselves and everyone who reads this code
    return func() error {
        // Welcome to nested function hell, population: you
        if err := func() error {
            // This is where sanity comes to die
            if err := func() error {
                // At this point we're just fucking with people
                if err := func() error {
                    // The beginning of the end
                    if err := func() error {
                        // Rock bottom of this shitshow
                        return ctxerrors.New("step 1 went to shit")
                    }(); err != nil {
                        // Step 2: electric boogaloo of failure
                        return ctxerrors.Wrap(err, "step 2 couldn't handle step 1's bullshit")
                    }
                    // If we somehow made it this far, we're lying
                    return nil
                }(); err != nil {
                    // Step 3: the reckoning
                    return ctxerrors.Wrap(err, "step 3 is having a mental breakdown")
                }
                // Still pretending everything is fine
                return nil
            }(); err != nil {
                // Step 4: fuck it, we're done trying
                return ctxerrors.Wrap(err, "step 4 said fuck this shit")
            }
            // The calm before the storm
            return nil
        }(); err != nil {
            // The final boss of this clusterfuck
            return ctxerrors.Wrap(err, "the entire fucking operation is fucked")
        }
        // Narrator: it was not fine
        return nil
    }() // Because we needed one more layer of stupid
}
```

Output when everything goes to hell, from Go 1.26 with the longest names cut short. The function names are what Go gives anonymous functions after the compiler inlines them into each other:

```
the entire fucking operation is fucked: step 4 said fuck this shit: step 3 is having a mental breakdown: step 2 couldn't handle step 1's bullshit: step 1 went to shit
[/path/to/main.go:50 in main.performStupidlyComplexOperation.performStupidlyComplexOperation.func1.(...).func4.func5]
[/path/to/main.go:53 in main.performStupidlyComplexOperation.performStupidlyComplexOperation.func1.(...).func3.func4]
[/path/to/main.go:59 in main.performStupidlyComplexOperation.performStupidlyComplexOperation.func1.performStupidlyComplexOperation.performStupidlyComplexOperation.func1.func2.func3]
[/path/to/main.go:65 in main.performStupidlyComplexOperation.performStupidlyComplexOperation.func1.func2]
[/path/to/main.go:71 in main.performStupidlyComplexOperation.func1]
```

This shit makes debugging actually bearable instead of wanting to throw your laptop out the fucking window.
