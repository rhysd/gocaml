#include <stdio.h>
#include <inttypes.h>
#include <stdlib.h>
#include "gocaml.h"

void print_int(gocaml_int const i)
{
    printf("%" PRId64, i);
}

void print_bool(gocaml_bool const i)
{
    printf("%s", i ? "true" : "false");
}

void print_float(gocaml_float const d)
{
    printf("%lg", d);
}

void println_int(gocaml_int const i)
{
    printf("%" PRId64 "\n", i);
}

void println_bool(gocaml_bool const i)
{
    printf("%s\n", i ? "true" : "false");
}

void println_float(gocaml_float const d)
{
    printf("%lg\n", d);
}

gocaml_int float_to_int(gocaml_float const f)
{
    return (gocaml_int) f;
}

gocaml_float float_of_int(gocaml_int const i)
{
    return (gocaml_float) i;
}

