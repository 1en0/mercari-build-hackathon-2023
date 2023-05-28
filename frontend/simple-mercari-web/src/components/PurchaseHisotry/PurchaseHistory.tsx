import React from "react";
import { ItemList } from "../ItemList";
import { useState, useEffect } from "react";
import { useCookies } from "react-cookie";
import { useParams } from "react-router-dom";
import { MerComponent } from "../MerComponent";
import { toast } from "react-toastify";
import { fetcher } from "../../helper";

interface Item {
	id: number;
	name: string;
	price: number;
	status: number;
	category_name: string;
}

export const PurchaseHistory: React.FC = () => {

	const [items, setItems] = useState<Item[]>([]);
	const [cookies] = useCookies(["token"]);
	const params = useParams();

	const fetchItems = () => {
		fetcher<Item[]>(`/users/${params.id}/purchase`, {
			method: "GET",
			headers: {
				"Content-Type": "application/json",
				Accept: "application/json",
				Authorization: `Bearer ${cookies.token}`,
			},
		})
			.then((items) => setItems(items))
			.catch((err) => {
				console.log(`GET error:`, err);
				toast.error(err.message);
			});
	};

	useEffect(() => {
		fetchItems();
	}, []);

	return (
		<MerComponent>
			<ItemList items={items} />
		</MerComponent>
	);
};
