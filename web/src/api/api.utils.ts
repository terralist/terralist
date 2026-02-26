import type {
  AxiosError,
  AxiosInstance,
  AxiosResponse,
  CreateAxiosDefaults,
  InternalAxiosRequestConfig
} from 'axios';
import { decodeError, type ErrorCode } from './api.errors';
import { transformKeys, snakeToCamel, camelToSnake } from '@/lib/conversions';
import axios from 'axios';

type ResultOK<T> = {
  status: 'OK';
  message: string;
  data: T;
  errors: [];
};

type ResultError = {
  status: 'ERROR';
  data?: never;
  message: string;
  errors: string[];
};

type Result<T> = ResultOK<T> | ResultError;

const withSuccess = <T>(data: T): ResultOK<T> => {
  return {
    status: 'OK',
    data: data,
    message: '',
    errors: []
  };
};

type Error = {
  errors: string[];
};

function isString(arg: unknown): arg is string {
  return typeof arg == 'string';
}

function isStringArray(arg: unknown): arg is string[] {
  return (
    typeof arg == 'object' &&
    Array.isArray(arg) &&
    arg.length > 0 &&
    isString(arg[0])
  );
}

function isError(arg: unknown): arg is Error {
  return (
    (arg as Error)?.errors != undefined && isStringArray((arg as Error).errors)
  );
}

const withError = (errorCode?: ErrorCode, data?: unknown): ResultError => {
  let errors: string[] = [];
  if (isError(data)) {
    errors = data.errors;
  } else if (isStringArray(data)) {
    errors = data;
  } else if (isString(data)) {
    errors = [data];
  } else {
    errors = [JSON.stringify(data)];
  }

  return {
    status: 'ERROR',
    message: decodeError(errorCode ?? 500),
    errors: errors
  };
};

const handleResponse = <T>(response: AxiosResponse<T>): Result<T> => {
  if (response.status >= 200 && response.status < 300) {
    return withSuccess<T>(response.data);
  }

  return withError(response.status, response.data);
};

const handleError = (error: AxiosError): ResultError => {
  if (error.response) {
    return withError(error.response?.status, error.response?.data);
  }

  if (error.message) {
    return withError(500, error.message);
  }

  return withError();
};

const responseConvertor = (response: AxiosResponse): AxiosResponse => {
  const { data, ...rest } = response;
  return { data: transformKeys(data, snakeToCamel), ...rest } as AxiosResponse;
};

const requestConvertor = (
  request: InternalAxiosRequestConfig
): InternalAxiosRequestConfig => {
  const { data, ...rest } = request;

  return {
    data: transformKeys(data, camelToSnake),
    ...rest
  } as InternalAxiosRequestConfig;
};

const createClient = (config?: CreateAxiosDefaults): AxiosInstance => {
  const client = axios.create(config);

  client.interceptors.request.use(requestConvertor);
  client.interceptors.response.use(responseConvertor);

  return client;
};

export {
  type Result,
  type ResultOK,
  type ResultError,
  handleResponse,
  handleError,
  createClient
};
