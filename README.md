# FXScript

A simple mission script language parser and runtime for Go.

## Features

- Labels
- Macros
- Pointers
- Expressions
- Preprocessor

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

// In the `ParserConfig` call, you can provide a filesystem for @include
// and a lookup function for @def directives.
script, err := fx.LoadScript([]byte("set health, 100\n"), vmConfig.ParserConfig(nil, nil))
```

### 3. Run the Script

```go
r := vm.NewRuntime(script, vmConfig)
myEnv := &MyEnvironment{values: make(map[fx.Identifier]int)}

r.Start(0, myEnv)
```

### 4. Hooks

You can use hooks to intercept command execution or argument unmarshalling.

```go
vmConfig := &vm.RuntimeConfig{
    Hooks: &vm.Hooks{
        PreExecute: func(cmd *fx.CommandNode) {
            fmt.Printf("Executing command: %s\n", cmd.Name)
        },
        PostExecute: func(cmd *fx.CommandNode, jumpPc int, jump bool) {
            fmt.Printf("Command executed: %s\n", cmd.Name)
        },
        PostUnmarshalArgs: func(args any) {
            fmt.Printf("Arguments unmarshalled: %+v\n", args)
        },
    },
}
```

### 6. Call a Label from Go

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

### Operators and Expressions

FXScript supports arithmetic, logical, and bitwise expressions.

| Category | Operators                                                           |
| :--- |:--------------------------------------------------------------------|
| **Arithmetic** | `+`, `-`, `*`, `/`, `%`                                             |
| **Bitwise** | `&` (AND), `\|` (OR), `^` (XOR), `<<` (LSH), `>>` (RSH)             |
| **Comparison** | `==`, `!=`, `<`, `>`, `<=`, `>=`                                    |
| **Unary** | `-` (negation), `*` (deref), `&` (addr), `^` (NOT), `!` (logic NOT) |

#### Pointer and Address Operators

- `&<value>`: Returns **address** of `<value>` if the `<value>` is an `identifier` or the `<value>` itself if it's an integer.
- `*<expr>`: Treats the result of `<expr>` as an address and returns the value from the memory at that address.

```
set a, 100
set b, *a     // b = value of memory at address 100
set c, *(a+1) // c = value of memory at address (a + 1)
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

- `def name value`: Script-level Define. Somewhat like a `#define` in C, but only for expressions.
- `var name`: Declares a script-level variable. The runtime will automatically assign an address to this variable.
- `var name[size]`: Declares a script-level array variable. `size` must be a static expression.
- `macro name ... endmacro`: Defines a macro. [See Macros](#macros) for details.
- `@include "file"`: Includes another file during preprocessing. Requires `fs.FS` to be provided in `ParserConfig`.
- `@def <argument>`: Can be used to inject anything as definition. Using the `LookupFn` provided in `ParserConfig`. The directive is replaced with `def <lookup return value>`.

### Macros

Macros allow you to define reusable blocks of code. They can also take parameters, which are prefixed with a `$` sign.

```
macro my_macro $param1, $param2
    set A, ($param1 + $param2)
endmacro

my_macro 10, 20
```

In this example, `my_macro 10, 20` will be expanded to `set A, (10 + 20)`.

Macro arguments are literally replaced in the macro body.

#### Macro Local Labels

Labels starting with a `%` (e.g., `%loop:`) are local to the macro they are defined in. When the macro is expanded, these labels are prefixed with a unique identifier to prevent name collisions if the macro is used multiple times.

```
macro my_loop $count
    set A, 0
    %loop:
        set A, (A + 1)
        jumpIf (A < $count), %loop
endmacro

my_loop 10
my_loop 20
```

### Built-in Commands

- `nop`: No operation.
- `set <ident>, <value>`: Sets identifier to value.
- `push <ident>`: Pushes a ident value onto the stack.
- `pop <ident>`: Pops a value from the stack and stores it at the address of the identifier.
- `goto <label/addr>`: Jumps to label or address.
- `call <label/addr>`: Calls subroutine at label or address.
- `ret`: Returns from subroutine.
- `jumpIf <condition>, <label/addr>`: Jumps to target if `<condition>` evaluates to a non-zero value.

### Array Variables

Array variables are declared using the `var` directive with a size in square brackets. The size must be an expression that can be evaluated at parse time.

```
def SIZE 3
var my_array[SIZE]

set my_array[0], 10
set my_array[1], 20
set my_array[2], 30

set A, my_array[1]
```

Arrays are stored in sequential memory addresses. You can use the address operator `&` to get the base address and use pointer arithmetic to access elements:

```
set &my_array + 1, 42 // same as set my_array[1], 42
```

Array indices can also be dynamic expressions:

```
set i, 2
set my_array[i], 100
```

## Custom Commands

You can extend FXScript with your own commands:

```go
const CmdMyCustom = fx.UserCommandOffset + 1

vmConfig := &vm.RuntimeConfig{
    UserCommands: []*vm.Command{
        {
            Name: "myCommand",
            Type:  CmdMyCustom,
            Handler: func(f *vm.Frame, args []fx.ExpressionNode) (jumpTarget int, jump bool) {
                fmt.Println("Custom command executed!")
                return
            },
        },
    },
}

script, err := fx.LoadScript([]byte("myCommand\n"), vmConfig.ParserConfig(nil, nil))

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
func(f *vm.Frame, args []fx.ExpressionNode) (jumpTarget int, jump bool)
```

- `jumpTarget`: The new Program Counter (PC) value if a jump should occur.
- `jump`: If `true`, the runtime will set the PC to `jumpTarget`. If `false`, the runtime continues with the next command.

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
        Type: CmdMyCustom,
        Handler: func(f *vm.Frame, args []fx.ExpressionNode) (jumpTarget int, jump bool) {
            return vm.WithArgs(f, args, func(f *vm.Frame, a *MyArgs) (jumpTarget int, jump bool) {
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

- `fx.Identifier`: 
    - If a plain identifier (like `health`) or `&health` is passed, it unmarshals to its **address**.
    - If an expression is passed (like `*health`), it is evaluated and cast to `fx.Identifier`.
- `int`, `float64`, `string`: Evaluates the expression and casts the result to the field type.

#### Example: The `set` command

The `set` command uses `Variable fx.Identifier` and `Value int`.

- `set A, 10`: Sets the memory at address `A` to `10`.
- `set A, B`: Sets memory `A` to the **value** of `B`.
- `set A, *B`: Sets memory `A` to the value of the variable pointed to by `B`.
- `set A, &B`: Sets memory `A` to the **address** of `B`.
- `set *A, 10`: Sets the memory value whose address is the value of `A` to `10`.
