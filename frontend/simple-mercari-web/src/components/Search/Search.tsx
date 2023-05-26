import { useState, useEffect } from "react";
import { toast } from "react-toastify";
import { fetcher } from "../../helper";

import {
	Grid,
	Select,
	MenuItem,
	TextField,
	IconButton,
} from "@mui/material";

import { Search } from '@mui/icons-material';

interface Item {
	id: number;
	name: string;
	price: number;
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
}

interface Prop {
	setItems: (items: Item[]) => void;
}

export const SearchFiled: React.FC<Prop> = (props) => {
	const [categories, setCategories] = useState<Category[]>([]);
	const [search, setSearch] = useState<SearchKey>({ category: -1, keyword: "", price_min: 1, price_max: 9999999 });

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

	const handleSubmit = () => {
		var query = ""
		if (search?.category)
			query = `/search_items?category=${search?.category}&keyword=${search?.keyword}&price_min=${search?.price_min}&price_max=${search?.price_max}`
		else
			query = `/search_items?category=${search?.category}&price_min=${search?.price_min}&price_max=${search?.price_max}`
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
	}, []);

	return (
		<div>
			<Grid container className="ItemSearch" spacing={2}>
				<Grid item>
					<span>
						<p>Category</p>
					</span>
					<Select
						name="category_id"
						id="MerTextInput"
						sx={{ display: "flex" }}
						className="SearchForm"
						label="category"
						defaultValue={-1}
						onChange={e => setSearch({ ...search, category: e.target.value })}
					>
						{categories &&
							categories.map((category) => {
								return <MenuItem value={category.id}>{category.name}</MenuItem>;
							})}
					</Select>
				</Grid>
				<Grid item>
					<span>
						<p>Keyword</p>
					</span>
					<TextField
						className="SearchForm"
						defaultValue=""
						onChange={e => setSearch({ ...search, keyword: e.target.value })}
					>検索</TextField>
				</Grid>
			</Grid>
			<span>
				<p>Price</p>
			</span>
			<Grid container className="ItemSearch" spacing={2}>
				<Grid item>
					<TextField
						className="SearchForm"
						defaultValue={1}
						onChange={e => setSearch({ ...search, price_min: parseInt(e.target.value) })}
					>min</TextField>
				</Grid>
				<Grid item>
					<div>-</div>
				</Grid>
				<Grid item>
					<TextField
						className="SearchForm"
						defaultValue={99999999}
						onChange={e => setSearch({ ...search, price_max: parseInt(e.target.value) })}
					>max</TextField>
				</Grid>
				<Grid item>
					<IconButton color="primary" onClick={handleSubmit}>
						<Search />
					</IconButton>
				</Grid>
			</Grid>
		</div>
	)
};
