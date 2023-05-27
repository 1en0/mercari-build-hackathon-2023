import { Login } from "../Login";
import { Signup } from "../Signup";
import { ItemList } from "../ItemList";
import { SearchFiled } from "../Search";
import { useCookies } from "react-cookie";
import { MerComponent } from "../MerComponent";
import { useState, useCallback } from "react";
import "react-toastify/dist/ReactToastify.css";


interface Item {
	id: number;
	name: string;
	price: number;
	status: number;
	category_name: string;
}

export const Home = () => {
	const [cookies] = useCookies(["userID", "token"]);
	const [items, setItems] = useState<Item[]>([]);



	const onSearchItem = useCallback(
		(items: Item[]) => {
			setItems(items);
		},
		[]
	);

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
				<SearchFiled setItems={onSearchItem}/>
				<ItemList items={items} />
			</div>
		</MerComponent>
	);

	return <>{cookies.token ? itemListPage : signUpAndSignInPage}</>;
};
