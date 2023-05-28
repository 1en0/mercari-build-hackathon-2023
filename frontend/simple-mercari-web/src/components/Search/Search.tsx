import { useState, useEffect } from "react";
import { toast } from "react-toastify";
import { fetcher } from "../../helper";

import {
	Grid,
	Select,
	MenuItem,
	TextField,
	IconButton,
	Checkbox,
} from "@mui/material";

import { Search } from '@mui/icons-material';

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
		<Grid className="ItemSearch">
			<Grid container spacing={2}>
				<Grid item width="35%">
					<span>
						<p>Category</p>
					</span>
					<Select
						name="category_id"
						id="MerTextInput"
						sx={{ display: "flex" }}
						className="SearchForm"
						label="category"
						size="small"
						defaultValue={-1}
						onChange={e => setSearch({ ...search, category: e.target.value })}
					>
						<MenuItem value={-1}>All</MenuItem>
						{categories &&
							categories.map((category) => {
								return <MenuItem value={category.id}>{category.name}</MenuItem>;
							})}
					</Select>
				</Grid>
				<Grid item width="35%">
					<span>
						<p>Keyword</p>
					</span>
					<TextField
						className="SearchForm"
						defaultValue=""
						size="small"
						onChange={e => setSearch({ ...search, keyword: e.target.value })}
					>検索</TextField>
				</Grid>
				<Grid item width="25%">
					<span>
						<p>Including Soldout</p>
					</span>
					<Checkbox
						className="SearchForm"
						color="primary"
						style={{backgroundColor: 'whitesmoke'}}
						onChange={() => setSearch({ ...search, is_include_soldout: !search.is_include_soldout })}
					/>
				</Grid>
			</Grid>
				<span>
					<p>Price</p>
				</span>
			<Grid container spacing={2}>
				<Grid item width="35%">
					<TextField
						className="SearchForm"
						defaultValue={1}
						size="small"
						fullWidth={false}
						type="number"
						onChange={e => setSearch({ ...search, price_min: parseInt(e.target.value) })}
					>min</TextField>
				</Grid>
				<Grid item width="5%">
					<div>-</div>
				</Grid>
				<Grid item width="35%">
					<TextField
						className="SearchForm"
						defaultValue={99999999}
						size="small"
						fullWidth={false}
						type="number"
						onChange={e => setSearch({ ...search, price_max: parseInt(e.target.value) })}
					>max</TextField>
				</Grid>
				<Grid item>
					<IconButton color="primary" size="large" onClick={handleSubmit}>
						<Search/>
					</IconButton>
				</Grid>
			</Grid>
		</Grid>
	)
};
