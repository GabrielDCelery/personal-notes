### Chapter 7

#### Writing simple shell scripts

Shell scripts tend to have the `.sh` extension but it is not mandatory. You can name your script either `myscript` or `myscript.sh` and both will work.

Shell scripts tend to have the interpreter put into the first line in the form of a `#!/bin/bash` (shebang) but again that is not mandatory but more of a convention.

The script has to have its executable permission set by `chmod +x <filename>` otherwise you won't be able to run it.

##### Executing and debugging shell scripts

For adding comments use the `#` symbol which will comment out the rest of the line.

```sh
#!/bin/bash

# This is a comment
ls -la # This is also  a comment
```

If you want to have some feedback at runtime of what yur program is doing the most common way is to use the `echo` command. There is alaso a trick where you can add `-x` in front of the filename while executing to print every command before executing it. For example:

```sh
$ bash -x myscript.sh

+ ls -la            <---- this line with the + sign is what is being executed
total 34634

```

#### Understnding shell variables

To specify a variable within a shell script do the following:

```sh
VARIABLE=value
CITY="Springfield"
PI=3.14
MYDATE=$(date) # This will execute date and assign the results to MYDATE
```

When working with special characters like `$`, `!`, `*` that can have an actual meaning in a shell script dependent on hpw we use them either the characters will be interpreted or just printed out as a string.

If we want characters to get interpreted literally we have to use either the escapa character `\` or wrap the section in a single quote `'`.

```sh
echo $HOME      # Will print /home/username
echo \$HOME     # Will print $HOME
echo '$HOME'    # Will print $HOME
```

Double quotes `"` are trickier. `$`, `!` will be interpreted but other special characters like `*` will be printed out literally.

#### Special shell parameters

Some shell parameters have special meanings and thei values get assigned automatically. Imagine you have a shell script like this:

```sh
#!/bin/bash
echo $0 # The name of the command
echo $1 # The first parameter
echo $2 # The second parameter
echo $# # Number of arguments given
echo $@ # All the arguments
echo $? # The exit status of the last command
```

When running the above you will get:

```sh
$ ./mayscript A B

A
B
2
A B
0
```

#### Prompting user for typing in parameters

You can use the `read` command to prompt the user to provide parameters. Important, the below snippet just collects the parameters if they were typed in but does not force the user to do so.

```sh
read param1 param2
echo $param1 $param2
```

#### Parameter expansion

One of the key aspects of shell variables is that `$VAR` is just a shorthand to `${VAR}`. This is important because if we use the long form of variables we can do some really neat tricks.
