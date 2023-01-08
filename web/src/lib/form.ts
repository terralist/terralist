import type { Validation, ValidationResult } from "./validation";

type InputType = 'email' | 'text' | 'textarea' | 'password' | 'number';

type FormEntry = {
  name: string,
  value?: string,
  type: InputType,
  required?: boolean,
  validations?: Validation<any>[],
};

const validateEntry = (entry: FormEntry): ValidationResult => {
  if (entry.required && !entry.value) {
    return {
      passed: false,
      message: "Required field.",
    } satisfies ValidationResult;
  }

  if (!entry.required && !entry.value) {
    return {
      passed: true,
    } satisfies ValidationResult;
  }

  for (let validation of entry.validations) {
    let result = validation(entry.value);

    if (!result.passed) {
      return {
        passed: false,
        message: result.message
      } satisfies ValidationResult
    }
  }

  return {
    passed: true
  } satisfies ValidationResult;
};

export {
  type InputType,
  type FormEntry,
  validateEntry,
};