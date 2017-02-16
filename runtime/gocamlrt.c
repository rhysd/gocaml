#include <stdio.h>
#include <stdint.h>
#include <stdlib.h>

void print_int(int64_t const i)
{
    printf("%lld", i);
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
    printf("%lld\n", i);
}

void println_bool(int const i)
{
    printf("%s\n", i ? "true" : "false");
}

void println_float(double const d)
{
    printf("%lg\n", d);
}

