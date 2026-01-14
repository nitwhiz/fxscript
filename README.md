# FXScript

A simple mission script language parser and runtime for Go.

## Features

- Custom commands
- Expression evaluation
- Identifiers and constants
- Labels and control flow (`goto`, `call`, `ret`)
- Preprocessor directives (`@include`)
- Script-level directives (`const`, `macro`)
- Conditional jumps (`jumpIf`, `jumpIfFlag`, `jumpIfNotFlag`)
- `hostCall` for `vm.Environment` interaction
- Pointers `*`

## Installation

```bash
go get -u github.com/nitwhiz/fxscript
```

## Basic Usage

To use FXScript, you need to:
1. Define your runtime environment by implementing the `vm.Environment` interface.
2. Configure the parser with your custom commands and variables.
3. Load and run your script.

### 1. Implement Runtime Environment

The `Environment` interface allows the `Runtime` to interact with your application.

```go
type MyEnvironment struct {
    values map[fx.Identifier]int
}

func (e *MyEnvironment) Get(variable fx.Identifier) int {
    return e.values[variable]
}

func (e *MyEnvironment) Set(variable fx.Identifier, value int) {
    e.values[variable] = value
}

func (e *MyEnvironment) HandleError(err error) {
    fmt.Printf("Runtime error: %v\n", err)
}

func (e *MyEnvironment) HostCall(f *vm.RuntimeFrame, args []any) (pc int, jump bool) {
    // Handle host calls from script
    return
}
```

### 2. Configure and Load Script

```go
config := &fx.ParserConfig{
    Variables: fx.IdentifierTable{
        "health": 1,
        "score":  2,
    },
}

script, err := fx.LoadScript([]byte("set health 100\n"), config)
```

### 3. Run the Script

```go
r := vm.NewRuntime(script)
myEnv := &MyEnvironment{values: make(map[fx.Identifier]int)}

r.Start(0, myEnv)
```

### 4. Call a Label from Go

You can also start execution from a specific label in your script:

```go
r.Call("myLabel", myEnv)
```

## Script Syntax

### Identifiers and Values

Identifiers are mapped to integer IDs. In the script, they are used by name if defined in the `ParserConfig`.

```
set health 100
add health 10
```

### Pointer and Invert Operators

Use `*` to retrieve a value from an address pointed to by an identifier or expression:

```
set a 100
set b (*a + 1)
```

Use `^` for bitwise NOT:

```
set a ^1
```

### Expressions

FXScript supports basic arithmetic expressions:

```
set score (10 + 20 * 2)
```

### Labels and Control Flow

```
set health 100
loop:
    add health -1
    jumpIf health 0 end
    goto loop
end:
    nop
```

### Preprocessor and Directives

- `@include "other.fx"`: Includes another script file
- `const name value`: Defines a constant
- `macro name ... endmacro`: Defines a macro

### Built-in Commands

- `nop`: No operation
- `set <ident> <value>`: Set identifier to value
- `copy <from_ident> <to_ident>`: Copy value from one identifier to another
- `add <ident> <value>`: Add value to identifier value in memory
- `goto <label/addr>`: Jump to label or address
- `call <label/addr>`: Call subroutine
- `ret`: Return from subroutine
- `jumpIf <ident> <value> <label/addr>`: Jump if identifier equals value
- `hostCall ...args`: Call `HostCall` on the runtime environment

## Custom Commands

You can extend FXScript with your own commands:

```go
const CmdMyCustom = fx.UserCommandOffset + 1

config := &fx.ParserConfig{
    CommandTypes: fx.CommandTypeTable{
        "myCommand": CmdMyCustom,
    },
}

script, err := fx.LoadScript([]byte("myCommand\n"), config)

if err != nil {
    panic(err)
}

r := vm.NewRuntime(script)

r.RegisterCommands([]*vm.Command{
    {
        Typ: CmdMyCustom,
        Handler: func(f *vm.RuntimeFrame, args []fx.ExpressionNode) (jumpTarget int, jump bool) {
            fmt.Println("Custom command executed!")
            return
        },
    },
})

myEnv := &MyEnvironment{values: make(map[fx.Identifier]int)}
r.Start(0, myEnv)
```

### Command Handlers and Return Types

The `Handler` function for a custom command has the following signature:

```go
func(f *vm.RuntimeFrame, args []fx.ExpressionNode) (jumpTarget int, jump bool)
```

- **`jumpTarget`**: The new Program Counter (PC) value if a jump should occur.
- **`jump`**: If `true`, the runtime will set the PC to `jumpTarget`. If `false`, the runtime continues with the next command.

#### Using `vm.WithArgs`

For commands that take arguments, you can use the `vm.WithArgs` helper to automatically unmarshal and evaluate arguments into a struct:

```go
type MyArgs struct {
    Target fx.Identifier `arg:""`
    Value  int         `arg:""`
}

r.RegisterCommands([]*vm.Command{
    {
        Typ: CmdMyCustom,
        Handler: func(f *vm.RuntimeFrame, args []fx.ExpressionNode) (jumpTarget int, jump bool) {
            return vm.WithArgs(f, args, func(f *vm.RuntimeFrame, a *MyArgs) (jumpTarget int, jump bool) {
                f.Set(a.Target, a.Value * 2)
                return
            })
        },
    },
})
```
