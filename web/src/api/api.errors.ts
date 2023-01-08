type ErrorCode = string | number;

type APIErrorsMap = {
  [key: string]: string
};

const apiErrorsMap: APIErrorsMap = {
  "INTERNAL_SERVER_ERROR": "Something went wrong internally. Please contact the platform administrator.",
  "UNKNOWN_ERROR": "Something wrong happened. Please, try again later."
};

const convertNumberCodeToString = (errorCode: number): string => {
  switch (errorCode) {
    case 500: return "INTERNAL_SERVER_ERROR";
    default: return "UNKNOWN_ERROR";
  }
};

const decodeError = (errorCode: ErrorCode): string => {
  if (typeof(errorCode) === "number") {
    errorCode = convertNumberCodeToString(errorCode);
  }

  return apiErrorsMap[errorCode] || apiErrorsMap["UNKNOWN_ERROR"];
};

export {
  type ErrorCode,
  decodeError
};