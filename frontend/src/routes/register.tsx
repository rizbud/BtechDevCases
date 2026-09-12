import {
  createFileRoute,
  Link,
  redirect,
  useNavigate,
} from "@tanstack/react-router";
import { useRegisterForm } from "@/auth/hooks/useRegisterForm";
import { PasswordInput } from "@/auth/components/PasswordInput";

export const Route = createFileRoute("/register")({
  beforeLoad: () => {
    if (localStorage.getItem("token")) {
      throw redirect({ to: "/" });
    }
  },
  component: Register,
});

function Register() {
  const navigate = useNavigate();
  const { register, errors, isSubmitting, isSuccess, onSubmit } =
    useRegisterForm();

  if (isSuccess) {
    return (
      <div className="flex justify-center items-center min-h-[80vh]">
        <div className="card w-full max-w-sm bg-base-100 shadow-xl p-6 gap-3 text-center">
          <h2 className="text-xl font-bold text-success">
            Registration complete!
          </h2>
          <p className="text-sm">You can now log in with your new account.</p>
          <button
            className="btn btn-primary mt-2"
            onClick={() => navigate({ to: "/login" })}
          >
            Go to Login
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="flex justify-center items-center min-h-[80vh]">
      <form
        onSubmit={onSubmit}
        className="card w-full max-w-sm bg-base-100 shadow-xl p-6 gap-3"
      >
        <h2 className="text-xl font-bold">Register</h2>

        {errors.root && (
          <p className="text-error text-sm">{errors.root.message}</p>
        )}

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
            placeholder="At least 8 characters"
            {...register("password")}
          />
          {errors.password && (
            <p className="text-error text-sm">{errors.password.message}</p>
          )}
        </div>

        <div className="flex flex-col gap-1">
          <label className="fieldset-label">Confirm Password</label>
          <PasswordInput
            placeholder="Re-enter your password"
            {...register("confirmPassword")}
          />
          {errors.confirmPassword && (
            <p className="text-error text-sm">
              {errors.confirmPassword.message}
            </p>
          )}
        </div>

        <button
          type="submit"
          className="btn btn-primary mt-2"
          disabled={isSubmitting}
        >
          {isSubmitting ? "Registering..." : "Register"}
        </button>

        <p className="text-sm text-center">
          Already have an account?{" "}
          <Link to="/login" className="link link-primary">
            Login
          </Link>
        </p>
      </form>
    </div>
  );
}
