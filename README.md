# FXScript

A simple mission script language parser and runtime for Go.

## Features

- Labels and control flow
- Pointers and Addresses
- Expressions
- Preprocessor directives

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

func (e *MyEnvironment) Get(variable fx.Identifier) (value int) {
    return e.values[variable]
}

func (e *MyEnvironment) Set(variable fx.Identifier, value int) {
    e.values[variable] = value
}

func (e *MyEnvironment) HandleError(err error) {
    fmt.Printf("Runtime error: %v\n", err)
}
```

### 2. Configure and Load Script

```go
vmConfig := &vm.RuntimeConfig{
    Identifiers: fx.IdentifierTable{
        "health": 1,
        "score":  2,
    },
}

script, err := fx.LoadScript([]byte("set health, 100\n"), vmConfig.ParserConfig())
```

### 3. Run the Script

```go
r := vm.NewRuntime(script, vmConfig)
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

Identifiers are mapped to integer addresses. In the script, they are used by name if defined in the `ParserConfig`.

```
set health, 100
set health, (health + 10)
```

Note: Commas are optional and can be used to separate command arguments for better readability.

### Operators and Expressions

FXScript supports arithmetic, logical, and bitwise expressions.

| Category | Operators                                                           |
| :--- |:--------------------------------------------------------------------|
| **Arithmetic** | `+`, `-`, `*`, `/`, `%`                                             |
| **Bitwise** | `&` (AND), `\|` (OR), `^` (XOR), `<<` (LSH), `>>` (RSH)             |
| **Comparison** | `==`, `!=`, `<`, `>`, `<=`, `>=`                                    |
| **Unary** | `-` (negation), `*` (deref), `&` (addr), `^` (NOT), `!` (logic NOT) |

#### Pointer and Address Operators

- `&<ident>`: Returns the **address** of the identifier.
- `*<expr>`: Treats the result of `<expr>` as an address and returns the value of the variable at that address.

```
set a, 100
set b, *a     // b = value of variable at address 100
set c, *(a+1) // c = value of variable at address (a + 1)
set d, &a     // d = address of identifier 'a'

set flags, (flags | 1)        // Set bit 0
set isSet, (flags & 1)        // Check bit 0
set score, (10 + 20 * 2)
jumpIf (health < 10), danger_label
```

### Labels and Control Flow

```
set health, 100
loop:
    set health, (health - 1)
    jumpIf (health == 0), end
    goto loop
end:
    nop
```

### Preprocessor and Directives

- `const name value`: Defines a script-level constant.
- `macro name ... endmacro`: Defines a macro.
- `@include "file"`: Includes another file during preprocessing.

### Built-in Commands

- `nop`: No operation.
- `set <ident>, <value>`: Sets identifier to value.
- `goto <label/addr>`: Jumps to label or address.
- `call <label/addr>`: Calls subroutine at label or address.
- `ret`: Returns from subroutine.
- `jumpIf <condition>, <label/addr>`: Jumps to target if `<condition>` evaluates to a non-zero value.

## Custom Commands

You can extend FXScript with your own commands:

```go
const CmdMyCustom = fx.UserCommandOffset + 1

vmConfig := &vm.RuntimeConfig{
    UserCommands: []*vm.Command{
        {
            Name: "myCommand",
            Typ:  CmdMyCustom,
            Handler: func(f *vm.RuntimeFrame, args []fx.ExpressionNode) (jumpTarget int, jump bool) {
                fmt.Println("Custom command executed!")
                return
            },
        },
    },
}

script, err := fx.LoadScript([]byte("myCommand\n"), vmConfig.ParserConfig())

if err != nil {
    panic(err)
}

r := vm.NewRuntime(script, vmConfig)

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

For commands that take arguments, you can use the `vm.WithArgs` helper to automatically unmarshal and evaluate arguments into a struct using reflection.

```go
type MyArgs struct {
    Target fx.Identifier `arg:""`         // Unmarshals the identifier address
    Value  int           `arg:""`         // Evaluates expression to int
    Scale  float64       `arg:"2,optional"` // Optional 3rd argument (index 2)
}

r.RegisterCommands([]*vm.Command{
    {
        Typ: CmdMyCustom,
        Handler: func(f *vm.RuntimeFrame, args []fx.ExpressionNode) (jumpTarget int, jump bool) {
            return vm.WithArgs(f, args, func(f *vm.RuntimeFrame, a *MyArgs) (jumpTarget int, jump bool) {
                // a.Target is the fx.Identifier (the address)
                // a.Value is the evaluated integer result
                f.Set(a.Target, a.Value * 2)
                return
            })
        },
    },
})
```

### Argument Types and Unmarshalling

The `vm.WithArgs` helper supports unmarshalling into a struct. The behavior depends on the field type:

- **`fx.Identifier`**: 
    - If a plain identifier (like `health`) or `&health` is passed, it unmarshals to its **address**.
    - If an expression is passed (like `*health`), it is evaluated and cast to `fx.Identifier`.
- **`int`, `float64`, `string`**: Evaluates the expression and casts the result to the field type.

#### Example: The `set` command

The `set` command uses `Variable fx.Identifier` and `Value int`.

- `set A, 10`: Sets the variable at address `A` to `10`.
- `set A, B`: Sets variable `A` to the **value** of `B`.
- `set A, *B`: Sets variable `A` to the value of the variable pointed to by `B`.
- `set A, &B`: Sets variable `A` to the **address** of `B`.
- `set *A, 10`: Sets the variable whose address is the value of `A` to `10`.
