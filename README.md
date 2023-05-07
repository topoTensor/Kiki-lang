# Kiki-lang is a hobby made imperative programming language

To use the language, compile the go files using run.sh for Linux and run.ps1 for Windows. Use the executable with 2 arguments, the file name and the intepretation type. If the interpretation type is 0, it will run normally, if it's 1 it will print additional messages about the interpretation process, such as value evaluations and other.

```Ruby
    # comments are written after cage sign
    a = 2           # variable declaration, possible value types are Number, Boolean, Array and String
    
    # the language supports basic arithmetics, boolean algebra and array operations
    a = 3*(2+2)/5       # arithmetics: +, -, *, /
    b = (a==2.4)&True   # Boolean algebra: True, False, & and, | or, ! not
    c = [1,2,3]

    c = c^[1,2,3]   # ^ operation appends array
    c = c^[a+2]     # to append a single value wich is not an array, first make it an array
    q = c~1         # array indexing operation: ~

    str = "abc"     # strings are just a Number Array, indexing them returns a Number, append operation works the same way
    str = str^"def"

    # Decision making and loops: ==, !=, <, >, <=, >=

    if str~0 == 97 then
        a = True
    end

    i=0
    while i < 10 do
        i=i+1
    end

    # Functions are pure, they can't modify outside variables and must always return a value, they can't be called outside of an expression, expect some pre-made functions. They are also first class and can be passed as an argument using the $ reference sign

    function add(a,b)
        return a+b
    end
    function do_func(f,x,y)
        return f(x,y)
    end

    a=do_func($add, 2,3)

    # Pre-made functions are:
    
    print(a, "alo", "\n") # prints any permissable value, doesn't print new line by default
    r = read()            # reads input from console and immidietly evaluates it
    # writing 2+2 in console will asign r to 4
    # it can be used inside any expression

    print(len(str))       # len returns length of the array or string
    byte("a")             # returns the ascii number of the character, can accept only one argument as a string
    if str~0 == byte("a") then
        print(str~0)
    end

    # panic("message")    # throws an error with the argument value, accepts only single argument

    a=file_read(str)      # returns the text inside the file, argument is the file name
    file_write("foo.txt", "text") # writes the second argument inside the first argument's file

    import("test.a")      # imports the file's text into the current file, replacing the import statement
    # it can be called as a statement anywhere

    # Example functions
    
    function fibonacci(n)
        if n <= 1 then
            return n
        end
        return fibonacci(n-1) + fibonacci(n-2)
    end

    function map(arr, f)
        newarr=[]
        i=0
        l=len(arr)
        while i < l do
            newarr=newarr^[f(arr~i)]
            i=i+1
        end
        return newarr
    end

```
