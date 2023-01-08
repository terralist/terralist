type ValidationResult = {
  passed: boolean,
  message?: string,
};

type Validation<T> = (value: T) => ValidationResult;

const StringMinimumLengthValidation = (min: number, exclusive: boolean = false): Validation<string> => {
  const errorMessage = `Minimum length should be ${exclusive ? "greater" : "greater or equal"} than ${min}.`
  
  return (value: string): ValidationResult => {
    if (exclusive) {
      return {
        passed: value.length > min,
        message: errorMessage,
      } satisfies ValidationResult;
    }

    return {
      passed: value.length >= min,
      message: errorMessage,
    } satisfies ValidationResult;
  };
};

const URLValidation = (): Validation<string> => {
  const regexPattern = new RegExp("^(?:(?:(?:https?|ftp):)?\/\/)(?:\S+(?::\S*)?@)?(?:(?!(?:10|127)(?:\.\d{1,3}){3})(?!(?:169\.254|192\.168)(?:\.\d{1,3}){2})(?!172\.(?:1[6-9]|2\d|3[0-1])(?:\.\d{1,3}){2})(?:[1-9]\d?|1\d\d|2[01]\d|22[0-3])(?:\.(?:1?\d{1,2}|2[0-4]\d|25[0-5])){2}(?:\.(?:[1-9]\d?|1\d\d|2[0-4]\d|25[0-4]))|(?:(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)(?:\.(?:[a-z\u00a1-\uffff0-9]-*)*[a-z\u00a1-\uffff0-9]+)*(?:\.(?:[a-z\u00a1-\uffff]{2,})))(?::\d{2,5})?(?:[/?#]\S*)?$");

  const errorMessage = "Not a valid URL."

  return (value: string) => {
    return {
      passed: regexPattern.test(value),
      message: errorMessage,
    } satisfies ValidationResult;
  };
};

export {
  type Validation,
  type ValidationResult,
  StringMinimumLengthValidation,
  URLValidation,
};