import { useEffect, useState } from "react";
import { useParams, useNavigate } from "react-router-dom";
import { useCookies } from "react-cookie";
import { MerComponent } from "../MerComponent";
import { toast } from "react-toastify";
import { fetcher, fetcherBlob } from "../../helper";

const ItemStatus = {
	ItemStatusInitial: 1,
	ItemStatusOnSale: 2,
	ItemStatusSoldOut: 3,
} as const;

type ItemStatus = typeof ItemStatus[keyof typeof ItemStatus];

interface Item {
	id: number;
	name: string;
	category_id: number;
	category_name: string;
	user_id: number;
	price: number;
	status: ItemStatus;
	description: string;
	views: number;
}

export const ItemDetail = () => {
	const navigate = useNavigate();
	const params = useParams();
	const [item, setItem] = useState<Item>();
	const [itemImage, setItemImage] = useState<Blob>();
	const [cookies] = useCookies(["token", "userID"]);

	const fetchItem = () => {
		fetcher<Item>(`/items-auth/${params.id}`, {
			method: "GET",
			headers: {
				Accept: "application/json",
				"Content-Type": "application/json",
				Authorization: `Bearer ${cookies.token}`,
			},
		})
			.then((res) => {
				console.log("GET success:", res);
				setItem(res);
			})
			.catch((err) => {
				console.log(`GET error:`, err);
				toast.error(err.message);
			});

		fetcherBlob(`/items/${params.id}/image`, {
			method: "GET",
			headers: {
				Accept: "application/json",
				"Content-Type": "application/json",
			},
		})
			.then((res) => {
				console.log("GET success:", res);
				setItemImage(res);
			})
			.catch((err) => {
				console.log(`GET error:`, err);
				toast.error(err.message);
			});
	};

	const onSubmit = () => {
		fetcher<Item[]>(`/purchase-v2/${params.id}`, {
			method: "POST",
			headers: {
				Accept: "application/json",
				"Content-Type": "application/json",
				Authorization: `Bearer ${cookies.token}`,
			},
			body: JSON.stringify({
				user_id: Number(cookies.userID),
			}),
		})
			.then((_) => window.location.reload())
			.catch((err) => {
				console.log(`POST error:`, err);
				toast.error(err.message);
			});
	};

	const Confirm = () => {
    const confirmBox = window.confirm(
      "Do you really want to buy this item?"
    )
    if (confirmBox === true) {
			onSubmit()
    }
  };

	useEffect(() => {
		fetchItem();
	}, []);

	return (
		<div className="ItemDetail">
			<MerComponent condition={() => item !== undefined}>
				{item && itemImage && (
					<div className="ItemDetailBlock">
						<img
							src={URL.createObjectURL(itemImage)}
							className="DetailImage"
							alt="item"
							onClick={() => navigate(`/item/${item.id}`)}
						/>
						<h4 color="gray">Information:</h4>
						<table>
							<tbody>
								<tr>
									<th>Item Name</th>
									<td>{item.name}</td>
								</tr>
								<tr>
									<th>Price</th>
									<td>{item.price}</td>
								</tr>
								<tr>
									<th>UserID</th>
									<td>{item.user_id}</td>
								</tr>
								<tr>
									<th>Category</th>
									<td>{item.category_name}</td>
								</tr>
								<tr>
									<th>Views</th>
									<td>{item.views}</td>
								</tr>
							</tbody>
						</table>
							<h4>Description:</h4>
							<div className="box">
								<p>{item.description}</p>
							</div>
						{item.status === ItemStatus.ItemStatusSoldOut ? (
							<button disabled={true} onClick={onSubmit} id="MerDisableButton">
								SoldOut
							</button>
						) :
							item.user_id !== parseInt(cookies.userID) ? (
								<button onClick={() => Confirm()} id="MerButton">
									Purchase
								</button>
							) :
								(
									<button
										onClick={() => navigate(`/item/${item.id}/edit`)}
										id="MerButton">
										Edit
									</button>
								)
						}
					</div>
				)}
			</MerComponent>
		</div>
	);
};
