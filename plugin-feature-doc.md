# Plugin API is simple and well-documented

The current plugin API has an overly large surface area, poor documentation,
and unclear names.  The objective is to improve this, focusing on the UX of
primary use-case: creating new service authenticators.

Aha Card: https://cyberark.aha.io/features/AAM-92

- [Objective](#objective)
- [Experience (Summary)](#experience-summary)
- [Experience (Detailed)](#experience-detailed)
- [Technical Details](#technical-details)
- [Testing](#testing)
- [Documentation](#documentation)
- [Open Questions](#open-questions)
- [Stories](#stories)
- [Future Work](#future-work)

### Objective

- Simplify Plugin API so authors of new service authenticators have a good
  experience which requires learning only concepts relevant to their goal.
- Write clear documentation to make the simplified API accessible to new
  contributors

### Experience (Summary)

The new plugin API will support writing new service authenticators only.  Only
a minimal interface will be required.

### Experience (Detailed)

Authors will need to implement only `PluginInfo` and a new
`ServiceAuthenticator` interface.

`APIVersion` will be moved into `PluginInfo`, and the `plugin_type` will be
explicitly specified:

```go
var PluginInfo = map[string]string{
  "api_version": "2.0",
  "plugin_types": "service_authenticator",
  "version":     "0.0.7",
  "id":          "test-plugin",
  "name":        "Test Plugin",
  "description": "Test plugin to demonstrate plugin functionality",
}
```

Implementing authenticators will look like:

```go
func ServiceAuthenticators() map[string]func(plugin_v2.AuthenticatorOptions) plugin_v2.Authenticator {
  return map[string]func(plugin_v2.AuthenticatorOptions) plugin_v2.Authenticator{
    "sample-authenticator-plugin": samplePlugin.NewAuthenticator,
  }
}
```

With the precise interface of an `Authenticator` TBD.

##### Assumptions

- The primary interest in plugins, at least at first, will be in writing new service authenticators.

##### Workflow

1. Write and compile one or more authenticators according to the above
1. Place the resulting `.so` files in the correct directory (currently `/usr/local/lib/secretless`)
1. Restart secretless

### Technical Details

- Per above, we'll refactor so that only the new, above described minimal API is public
- All the old interfaces will be moved into internal
- We'll have a generic type doing what's currently repeated by all listeners
- We'll need to deal with the exceptions represented by the http handler --
  what we were calling "sub-handlers" in previous discussions (as a placeholder
  name).
- Once handlers and listeners are combined, go back and update our config code
  to use them directly.

### Testing

**Integration**

- A new example plugin to replace current one 
- Tests showing it loads and can be used successfully as authenticator specified in a v2 secretless.yml

**Unit**

A major aim of this refactor will be to have components that are unit testable.
Thus every new nontrivial type or func created will likely have unit tests.
What those types are is still TBD.

### Open Questions

- Per above, the "sub handler" question
- Definitive list of all places in `plugin.Manager` and elsewhere that we'll need to update
- API of `ServiceAuthenticator` interface
- Design of the generic "listener engine"

### Stories

The epic has links to all current stories:

- https://app.zenhub.com/workspaces/secretless-5c9d073c270ed03f5cb98c95/issues/cyberark/secretless-broker/693

### Future Work

- Further related refactoring of the plugin.Manager, Proxy, etc.
