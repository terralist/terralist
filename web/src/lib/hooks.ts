import { writable, type Readable, type Writable } from 'svelte/store';
import { onMount } from 'svelte';
import type { Result } from '@/api/api.utils';

const useFlag: (
  value: boolean,
  onSet?: () => void,
  onReset?: () => void
) => [Writable<boolean>, () => void, () => void] = (
  value = false,
  onSet?,
  onReset?
) => {
  const flag = writable(value);

  const enableFlag = (): void => {
    onSet?.();
    flag.set(true);
  };

  const disableFlag = (): void => {
    onReset?.();
    flag.set(false);
  };

  return [flag, enableFlag, disableFlag];
};

const useToggle: (
  value: boolean,
  onSet?: () => void,
  onReset?: () => void
) => [Writable<boolean>, () => void] = (value = false, onSet?, onReset?) => {
  let t: boolean = value;

  const [toggle, set, reset] = useFlag(t, onSet, onReset);

  const flip = (): void => {
    t = !t;
    (t ? set : reset)();
  };

  return [toggle, flip];
};

type QueryLoading = {
  data?: never;
  isLoading: true;
  error?: never;
};

type QueryOK<T> = {
  data: T;
  isLoading: false;
  error?: never;
};

type QueryError = {
  data?: never;
  isLoading: false;
  error: string;
};

type QueryResult<T> = QueryLoading | QueryOK<T> | QueryError;

// eslint-disable-next-line @typescript-eslint/no-explicit-any
type QueryFn<T> = (...args: any[]) => Promise<Result<T>>;

const useQuery: <T>(
  query: QueryFn<T>,
  ...args: any[] // eslint-disable-line @typescript-eslint/no-explicit-any
) => Readable<QueryResult<T>> = <T>(
  query: QueryFn<T>,
  ...args: any[] // eslint-disable-line @typescript-eslint/no-explicit-any
) => {
  const result: Writable<QueryResult<T>> = writable({
    isLoading: true
  } as QueryLoading);

  onMount(async () => {
    const { data: content, status, message } = await query(...args);

    if (status == 'ERROR') {
      result.set({ isLoading: false, error: message } as QueryError);
      return;
    }

    result.set({ data: content, isLoading: false } as QueryOK<T>);
  });

  return result;
};

export { useFlag, useToggle, useQuery };
