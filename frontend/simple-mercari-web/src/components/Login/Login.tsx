import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useCookies } from "react-cookie";
import { toast } from "react-toastify";
import { fetcher } from "../../helper";

export const Login = () => {
  const [userID, setUserID] = useState<number>();
  const [password, setPassword] = useState<string>();
  const [_, setCookie] = useCookies(["userID", "token"]);

  const navigate = useNavigate();

  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
	  event.preventDefault();
    fetcher<{ id: number; name: string; token: string }>(`/login`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        user_id: userID,
        password: password,
      }),
    })
      .then((user) => {
        toast.success("Signed in!");
        console.log("POST success:", user.id);
        setCookie("userID", user.id);
        setCookie("token", user.token);
        navigate("/");
      })
      .catch((err) => {
        console.log(`POST error:`, err);
        toast.error(err.message);
      });
  };

  return (
    <div>
      <form onSubmit={onSubmit} className="Login">
        <label id="MerInputLabel">User ID</label>
        <input
          type="number"
          name="userID"
          id="MerTextInput"
          placeholder="UserID"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            setUserID(Number(e.target.value));
          }}
          required
					min={0}
        />
        <label id="MerInputLabel">Password</label>
        <input
          type="password"
          name="password"
          id="MerTextInput"
          placeholder="password"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            setPassword(e.target.value);
          }}
          required
					minLength={8}
        />
        <button type="submit" id="MerButton">
          Login
        </button>
      </form>
    </div>
  );
};
