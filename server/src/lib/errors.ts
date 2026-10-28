import { Elysia } from "elysia";

export type ErrorCode = "VALIDATION" | "NOT_FOUND" | "RATE_LIMITED" | "INTERNAL";

export type ErrorBody = {
  error: { code: ErrorCode; message: string; fieldErrors?: Record<string, string[]> };
};

/**
 * Throw from any handler to return `{ error: { code, message } }` with the given status.
 * (Named `errorCode`, not `code`: Elysia uses an error's `code` property to classify it.)
 */
export class HttpError extends Error {
  constructor(
    readonly status: number,
    readonly errorCode: ErrorCode,
    message: string,
  ) {
    super(message);
  }
}

function body(code: ErrorCode, message: string, fieldErrors?: Record<string, string[]>): ErrorBody {
  return { error: fieldErrors ? { code, message, fieldErrors } : { code, message } };
}

export const errorHandler = new Elysia({ name: "error-handler" })
  .error({ HttpError })
  .onError({ as: "global" }, ({ code, error, set }) => {
    switch (code) {
      case "HttpError":
        set.status = error.status;
        return body(error.errorCode, error.message);
      case "VALIDATION": {
        const fieldErrors: Record<string, string[]> = {};
        for (const issue of error.all) {
          const key = ("path" in issue && issue.path) || "_";
          (fieldErrors[key] ??= []).push(issue.message ?? "Invalid value");
        }
        set.status = 422;
        return body("VALIDATION", "Invalid request", fieldErrors);
      }
      case "PARSE":
        set.status = 400;
        return body("VALIDATION", "Malformed request body");
      case "NOT_FOUND":
        set.status = 404;
        return body("NOT_FOUND", "Route not found");
      default:
        console.error(error);
        set.status = 500;
        return body("INTERNAL", "Internal server error");
    }
  });
