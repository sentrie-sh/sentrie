import { something } from "@local/something";

export function User() {
  return {
    something: something(),
    name: "John Doe",
    email: "john.doe@example.com",
    role: "admin",
    status: "active",
  }
}