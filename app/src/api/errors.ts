// One entry of a 400 invalid_body response: which field broke which rule,
// using the JSON field names.
export type FieldError = { field: string; rule: string };

// The API answered with an error status. `code` is what the app reacts to and
// picks copy by; `message` is for developers and is never shown.
export class ApiError extends Error {
  readonly status: number;
  readonly code: string;
  readonly fields: FieldError[];

  constructor(
    status: number,
    code: string,
    message: string,
    fields: FieldError[] = [],
  ) {
    super(message);
    this.name = "ApiError";
    this.status = status;
    this.code = code;
    this.fields = fields;
  }
}

// The request got no response at all: no connection, the server unreachable,
// or the request timed out.
export class NetworkError extends Error {
  readonly timedOut: boolean;

  constructor(message: string, { timedOut = false } = {}) {
    super(message);
    this.name = "NetworkError";
    this.timedOut = timedOut;
  }
}
