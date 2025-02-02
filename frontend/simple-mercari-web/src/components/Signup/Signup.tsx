import React, { useState } from "react";
import { useNavigate } from "react-router-dom";
import { useCookies } from "react-cookie";
import { toast } from "react-toastify";
import { fetcher } from "../../helper";

export const Signup = () => {
  const [name, setName] = useState<string>();
  const [password, setPassword] = useState<string>();
  const [userID, setUserID] = useState<number>();
  const [_, setCookie] = useCookies(["userID"]);

  const navigate = useNavigate();
  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
	  event.preventDefault();
    fetcher<{ id: number; name: string }>(`/register`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        name: name,
        password: password,
      }),
    })
      .then((user) => {
        toast.success("New account is created!");
        console.log("POST success:", user.id);
        setCookie("userID", user.id);
        setUserID(user.id);
        navigate("/");
      })
      .catch((err) => {
        console.log(`POST error:`, err);
        toast.error(err.message);
      });
  };

  return (
    <div>
		 <form onSubmit={onSubmit} className="Signup">
        <label id="MerInputLabel">User Name</label>
        <input
          type="text"
          name="name"
          id="MerTextInput"
          placeholder="name"
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => {
            setName(e.target.value);
          }}
          required
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
          Signup
        </button>
        {userID ? (
          <p>Use "{userID}" as UserID for login</p>
        ) : null}
			</form>
    </div>
  );
};
