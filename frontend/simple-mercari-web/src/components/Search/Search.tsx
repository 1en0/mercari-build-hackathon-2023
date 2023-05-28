import { useState, useEffect } from "react";
import { toast } from "react-toastify";
import { fetcher } from "../../helper";

interface Item {
	id: number;
	name: string;
	price: number;
	status: number;
	category_name: string;
}

interface Category {
	id: number;
	name: string;
}

interface SearchKey {
	category: string | number;
	keyword: string;
	price_min: number;
	price_max: number;
	is_include_soldout: boolean;
}

interface Prop {
	setItems: (items: Item[]) => void;
}

export const SearchFiled: React.FC<Prop> = (props) => {
	const [categories, setCategories] = useState<Category[]>([]);
	const [search, setSearch] = useState<SearchKey>({ category: -1, keyword: "", price_min: 1, price_max: 9999999, is_include_soldout: false });

	const fetchCategories = () => {
		fetcher<Category[]>(`/items/categories`, {
			method: "GET",
			headers: {
				"Content-Type": "application/json",
				Accept: "application/json",
			},
		})
			.then((items) => setCategories(items))
			.catch((err) => {
				console.log(`GET error:`, err);
				toast.error(err.message);
			});
	};

	const fetchItems = () => {
		fetcher<Item[]>(`/search-detail`, {
			method: "GET",
			headers: {
				"Content-Type": "application/json",
				Accept: "application/json",
			},
		})
			.then((data) => {
				console.log("GET success:", data);
				props.setItems(data);
			})
			.catch((err) => {
				console.log(`GET error:`, err);
				toast.error(err.message);
			});
	};

	const handleSubmit = () => {
		var query = ""
		if (search?.category !== -1)
			query = `/search-detail?category=${search?.category}&name=${search?.keyword}&price-min=${search?.price_min}&price-max=${search?.price_max}&is-include-soldout=${search.is_include_soldout}`
		else
			query = `/search-detail?name=${search?.keyword}&price-min=${search?.price_min}&price-max=${search?.price_max}&is-include-soldout=${search.is_include_soldout}`
		fetcher<Item[]>(query, {
			method: "GET",
			headers: {
				"Content-Type": "application/json",
				Accept: "application/json",
			},
		})
			.then((data) => {
				console.log("GET success:", data);
				props.setItems(data);
			})
			.catch((err) => {
				console.log(`GET error:`, err);
				toast.error(err.message);
			});
	};

	useEffect(() => {
		fetchCategories();
		fetchItems();
	}, []);

	return (
		<div>
			<div>
				<div>
					<span>
						<p>Category</p>
					</span>
					<select
						name="category_id"
						id="MerTextInput"
						className="SearchForm"
						defaultValue={-1}
						onChange={e => setSearch({ ...search, category: e.target.value })}
					>
						<option value={-1}>All</option>
						{categories &&
							categories.map((category) => {
								return <option value={category.id}>{category.name}</option>;
							})}
					</select>
				</div>
				<div>
					<span>
						<p>Keyword</p>
					</span>
					<input
						className="SearchForm"
						defaultValue=""
						type="text"
						onChange={e => setSearch({ ...search, keyword: e.target.value })}
					/>
				</div>
				<div>
					<span>
						<p>Including Soldout</p>
					</span>
					<input
						type="checkbox"
						className="SearchForm"
						color="primary"
						style={{backgroundColor: 'whitesmoke'}}
						onChange={() => setSearch({ ...search, is_include_soldout: !search.is_include_soldout })}
					/>
				</div>
			</div>
				<span>
					<p>Price</p>
				</span>
			<div>
				<div>
					<input
						className="SearchForm"
						defaultValue={1}
						type="number"
						onChange={e => setSearch({ ...search, price_min: parseInt(e.target.value) })}
					/>
				</div>
				<div>
					<div>-</div>
				</div>
				<div>
					<input
						className="SearchForm"
						defaultValue={99999999}
						type="number"
						onChange={e => setSearch({ ...search, price_max: parseInt(e.target.value) })}
					/>
				</div>
				<div>
					<button color="primary" onClick={handleSubmit}>
							submit			
					</button>
				</div>
			</div>
		</div>
	)
};
