import type { AxiosError, AxiosInstance, AxiosResponse, CreateAxiosDefaults, InternalAxiosRequestConfig } from "axios";
import { decodeError, type ErrorCode } from "./api.errors";
import { transformKeys, snakeToCamel, camelToSnake } from "@/lib/conversions";
import axios from "axios";

interface Result<T> {
  status: 'OK' | 'ERROR',
  message?: string,
  data?: T,
};

const withSuccess = <T>(data: T): Result<T> => {
  return {
    status: 'OK',
    data: data,
    message: undefined,
  } satisfies Result<T>;
};

const withError = (errorCode: ErrorCode = undefined, data?: any): Result<undefined> => {
  if (data) {
    console.log("API Response:", errorCode, data);
  }

  return {
    status: 'ERROR',
    message: decodeError(errorCode),
  } satisfies Result<undefined>;
};

const handleResponse = <T>(response: AxiosResponse<T>): Result<T> => {
  if ([200, 201].includes(response.status)) {
    return withSuccess<T>(response.data);
  }

  return withError(response.status);
}

const handleError = (error: AxiosError): Result<undefined> => withError(error.response?.status, error.response?.data);

const responseConvertor = (response: AxiosResponse): AxiosResponse => {
  let { data, ...rest } = response;

  data = transformKeys(data, snakeToCamel);

  return { data, ...rest } as AxiosResponse;
}

const requestConvertor = (request: InternalAxiosRequestConfig): InternalAxiosRequestConfig => {
  let { data, ...rest } = request;

  data = transformKeys(data, camelToSnake);

  return { data, ...rest } as InternalAxiosRequestConfig;
};

const createClient = (config?: CreateAxiosDefaults<any>): AxiosInstance => {
  const client = axios.create(config);

  client.interceptors.request.use(requestConvertor);
  client.interceptors.response.use(responseConvertor);

  return client;
}

export {
  type Result,
  handleResponse,
  handleError,
  createClient,
};