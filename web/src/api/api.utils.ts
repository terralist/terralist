import { decodeError, type ErrorCode } from "./api.errors";

type APIResult<T> = {
  status: 'OK' | 'ERROR',
  message?: string,
  data?: T,
};

const successAPIResult = <T>(data: T): APIResult<T> => {
  return {
    status: 'OK',
    data: data,
  } satisfies APIResult<T>;
};

const errorAPIResult = (errorCode: ErrorCode = undefined): APIResult<any> => {
  return {
    status: 'ERROR',
    message: decodeError(errorCode),
  } satisfies APIResult<any>;
};

export type { APIResult };

export {
  successAPIResult,
  errorAPIResult,
};