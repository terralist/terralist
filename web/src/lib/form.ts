import type { Validation, ValidationResult } from './validation';

type InputType = 'email' | 'text' | 'textarea' | 'password' | 'number';

type FormEntry = {
  id: string;
  name: string;
  value?: string;
  type: InputType;
  required?: boolean;
  disabled?: boolean;
  // eslint-disable-next-line @typescript-eslint/no-explicit-any
  validations?: Validation<any>[];
};

const validateEntry = (entry: FormEntry): ValidationResult => {
  if (entry.required && !entry.value) {
    return {
      passed: false,
      message: 'This field is required.'
    } satisfies ValidationResult;
  }

  if (!entry.required && !entry.value) {
    return {
      passed: true,
      message: ''
    } satisfies ValidationResult;
  }

  if (entry.validations) {
    for (const validation of entry.validations) {
      const result = validation(entry.value);

      if (!result.passed) {
        return {
          passed: false,
          message: result.message
        } satisfies ValidationResult;
      }
    }
  }

  return {
    passed: true,
    message: ''
  } satisfies ValidationResult;
};

export { type InputType, type FormEntry, validateEntry };
