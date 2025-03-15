import { readable, writable, type Readable, type Writable } from "svelte/store";
import { onMount } from "svelte";
import type { Result } from "@/api/api.utils";

const useFlag: (
  value: boolean,
  onSet?: () => void,
  onReset?: () => void,
) => [Writable<boolean>, () => void, () => void] = (
  value = false,
  onSet?,
  onReset?,
) => {
  let flag = writable(value);

  const enableFlag = () => {
    onSet?.();
    flag.set(true);
  };

  const disableFlag = () => {
    onReset?.();
    flag.set(false);
  };

  return [flag, enableFlag, disableFlag];
};

const useToggle: (
  value: boolean,
  onSet?: () => void,
  onReset?: () => void,
) => [Writable<boolean>, () => void] = (value = false, onSet?, onReset?) => {
  let t: boolean = value;

  const [toggle, set, reset] = useFlag(t, onSet, onReset);

  const flip = () => {
    t = !t;
    (t ? set : reset)();
  };

  return [toggle, flip];
}

type QueryLoading = {
  data?: never;
  isLoading: true;
  error?: never;
}

type QueryOK<T> = {
  data: T;
  isLoading: false;
  error?: never;
}

type QueryError = {
  data?: never;
  isLoading: false;
  error: string;
}

type QueryResult<T> = QueryLoading | QueryOK<T> | QueryError;

const useQuery: <T>(
  query: (...args: any[]) => Promise<Result<T>>,
  ...args: any[]
) => Readable<QueryResult<T>> = <T>(query: (...args: any[]) => Promise<Result<T>>, ...args: any[]) => {
  let result: Writable<QueryResult<T>> = writable({ isLoading: true } as QueryLoading);

  onMount(async () => {
    const { data: content, status, message } = await query(...args);

    if (status == "ERROR") {
      result.set({ isLoading: false, error: message } as QueryError);
      return;
    }

    result.set({ data: content, isLoading: false } as QueryOK<T>);
  });

  return result;
};

export {
  useFlag,
  useToggle,
  useQuery,
};
