#include <stdio.h>
#include <stdint.h>
#include <inttypes.h>
#include <stdlib.h>

void print_int(int64_t const i)
{
    printf("%" PRId64, i);
}

void print_bool(int const i)
{
    printf("%s", i ? "true" : "false");
}

void print_float(double const d)
{
    printf("%lg", d);
}

void println_int(int64_t const i)
{
    printf("%" PRId64 "\n", i);
}

void println_bool(int const i)
{
    printf("%s\n", i ? "true" : "false");
}

void println_float(double const d)
{
    printf("%lg\n", d);
}

int64_t float_to_int(double const f)
{
    return (int64_t) f;
}

double float_of_int(int64_t const i)
{
    return (double) i;
}

