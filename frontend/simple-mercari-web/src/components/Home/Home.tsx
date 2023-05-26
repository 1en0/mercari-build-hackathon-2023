import { Login } from "../Login";
import { Signup } from "../Signup";
import { ItemList } from "../ItemList";
import { SearchFiled } from "../Search";
import { useCookies } from "react-cookie";
import { MerComponent } from "../MerComponent";
import { useEffect, useState, useCallback } from "react";
import { toast } from "react-toastify";
import { fetcher } from "../../helper";
import "react-toastify/dist/ReactToastify.css";


interface Item {
	id: number;
	name: string;
	price: number;
	category_name: string;
}

export const Home = () => {
	const [cookies] = useCookies(["userID", "token"]);
	const [items, setItems] = useState<Item[]>([]);

	const fetchItems = () => {
		fetcher<Item[]>(`/items`, {
			method: "GET",
			headers: {
				"Content-Type": "application/json",
				Accept: "application/json",
			},
		})
			.then((data) => {
				console.log("GET success:", data);
				setItems(data);
			})
			.catch((err) => {
				console.log(`GET error:`, err);
				toast.error(err.message);
			});
	};


	const onSearchItem = useCallback(
		(items: Item[]) => {
			setItems(items);
		},
		[]
	);


	useEffect(() => {
		fetchItems();
	}, []);

	const signUpAndSignInPage = (
		<>
			<div>
				<Signup />
			</div>
			or
			<div>
				<Login />
			</div>
		</>
	);

	const itemListPage = (
		<MerComponent>
			<div>
				<span>
					<p>Logined User ID: {cookies.userID}</p>
				</span>
				<SearchFiled setItems={onSearchItem} />
				<ItemList items={items} />
			</div>
		</MerComponent>
	);

	return <>{cookies.token ? itemListPage : signUpAndSignInPage}</>;
};
