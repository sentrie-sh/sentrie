// SPDX-FileCopyrightText: © 2026 Binaek Sarkar <binaek89@gmail.com>
// SPDX-License-Identifier: Apache-2.0

// Math functions
export const round = Math.round;
export const floor = Math.floor;
export const ceil = Math.ceil;
export const max = Math.max;
export const min = Math.min;
export const abs = Math.abs;
export const sqrt = Math.sqrt;
export const pow = Math.pow;
export const exp = Math.exp;
export const log = Math.log;
export const log10 = Math.log10;
export const log2 = Math.log2;
export const sin = Math.sin;
export const cos = Math.cos;
export const tan = Math.tan;
export const asin = Math.asin;
export const acos = Math.acos;
export const atan = Math.atan;
export const atan2 = Math.atan2;
export const sinh = Math.sinh;
export const cosh = Math.cosh;
export const tanh = Math.tanh;
export const random = Math.random;

// Math constants
export const E = Math.E;
export const PI = Math.PI;
export const LN2 = Math.LN2;
export const LN10 = Math.LN10;
export const LOG2E = Math.LOG2E;
export const LOG10E = Math.LOG10E;
export const SQRT2 = Math.SQRT2;
export const SQRT1_2 = Math.SQRT1_2;
export const MAX_VALUE = Number.MAX_VALUE;
export const MIN_VALUE = Number.MIN_VALUE;

// String functions
export const length = (str: string) => str.length;
export const fromCharCode = String.fromCharCode;
export const replaceAll = (str: string, search: string, replace: string) => str.replaceAll(search, replace);

// Number functions
export const isNaN = Number.isNaN;
export const parseInt = Number.parseInt;
export const parseFloat = Number.parseFloat;
export const isFinite = Number.isFinite;
export const isInteger = Number.isInteger;

// Date functions
// There is deliberately no `now` export: the Date clock reads the wall clock
// in milliseconds on every call. Use `now` from @sentrie/time, which is pinned to
// the execution start and returns seconds (#207).
export const dateParse = Date.parse;
export const UTC = Date.UTC;

// JSON functions
export const parse = JSON.parse;
export const stringify = JSON.stringify;

// Array functions
export const isArray = Array.isArray;
export const from = Array.from;
export const of = Array.of;

