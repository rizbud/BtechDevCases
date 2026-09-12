import { createFileRoute, Link, redirect } from "@tanstack/react-router";
import { useLoginForm } from "@/auth/hooks/useLoginForm";
import { PasswordInput } from "@/auth/components/PasswordInput";

export const Route = createFileRoute("/login")({
  beforeLoad: () => {
    if (localStorage.getItem("token")) {
      throw redirect({ to: "/" });
    }
  },
  component: Login,
});

function Login() {
  const { register, errors, isSubmitting, onSubmit } = useLoginForm();

  return (
    <div className="flex justify-center items-center min-h-[80vh]">
      <form
        onSubmit={onSubmit}
        className="card w-full max-w-sm bg-base-100 shadow-xl p-6 gap-3"
      >
        <h2 className="text-xl font-bold">Login</h2>

        <div className="flex flex-col gap-1">
          <label className="fieldset-label">Email</label>
          <input
            type="email"
            placeholder="you@example.com"
            className="input input-bordered w-full"
            {...register("email")}
          />
          {errors.email && (
            <p className="text-error text-sm">{errors.email.message}</p>
          )}
        </div>

        <div className="flex flex-col gap-1">
          <label className="fieldset-label">Password</label>
          <PasswordInput
            placeholder="Enter your password"
            {...register("password")}
          />
          {errors.password && (
            <p className="text-error text-sm">{errors.password.message}</p>
          )}
        </div>

        {errors.root && (
          <p className="text-error text-sm">{errors.root.message}</p>
        )}

        <button
          type="submit"
          className="btn btn-primary mt-2"
          disabled={isSubmitting}
        >
          {isSubmitting ? "Logging in..." : "Login"}
        </button>

        <p className="text-sm text-center">
          No account?{" "}
          <Link to="/register" className="link link-primary">
            Register
          </Link>
        </p>
      </form>
    </div>
  );
}
