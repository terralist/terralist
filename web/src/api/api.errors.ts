type ErrorCode = string | number;

type APIErrorsMap = {
  [key: string]: string;
};

const apiErrorsMap: APIErrorsMap = {
  BAD_REQUEST: 'Your request was not formatted properly.',
  FORBIDDEN: 'You do not have permission to access this resource.',
  NOT_FOUND: "The resource you're looking for was not found on the server.",
  INTERNAL_SERVER_ERROR:
    'Something went wrong internally. Please contact the platform administrator.',
  UNKNOWN_ERROR: 'Something wrong happened. Please, try again later.'
};

const convertNumberCodeToString = (errorCode: number): string => {
  switch (errorCode) {
    case 400:
      return 'BAD_REQUEST';
    case 403:
      return 'FORBIDDEN';
    case 404:
      return 'NOT_FOUND';
    case 500:
      return 'INTERNAL_SERVER_ERROR';
    default:
      return 'UNKNOWN_ERROR';
  }
};

const decodeError = (errorCode: ErrorCode): string => {
  if (typeof errorCode === 'number') {
    errorCode = convertNumberCodeToString(errorCode);
  }

  return apiErrorsMap[errorCode] || apiErrorsMap['UNKNOWN_ERROR'];
};

const FORBIDDEN_ERROR = apiErrorsMap['FORBIDDEN'];

export { type ErrorCode, FORBIDDEN_ERROR, decodeError };
