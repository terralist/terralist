import type { Validation, ValidationResult } from "./validation";

type InputType = 'email' | 'text' | 'textarea' | 'password' | 'number';

type FormEntry = {
  id: string,
  name: string,
  value?: string,
  type: InputType,
  required?: boolean,
  disabled?: boolean,
  validations?: Validation<any>[],
};

const validateEntry = (entry: FormEntry): ValidationResult => {
  if (entry.required && !entry.value) {
    return {
      passed: false,
      message: "This field is required.",
    } satisfies ValidationResult;
  }

  if (!entry.required && !entry.value) {
    return {
      passed: true,
    } satisfies ValidationResult;
  }

  if (entry.validations) {
    for (let validation of entry.validations) {
      let result = validation(entry.value);

      if (!result.passed) {
        return {
          passed: false,
          message: result.message
        } satisfies ValidationResult
      }
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