import type { AxiosError, AxiosResponse } from "axios";
import { decodeError, type ErrorCode } from "./api.errors";

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

const withError = (errorCode: ErrorCode = undefined): Result<undefined> => {
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

const handleError = (error: AxiosError): Result<undefined> => withError(error.response?.status);

export {
  type Result,
  handleResponse,
  handleError,
};