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
}

interface FormDataType {
	id: number;
	name: string;
	category_id: number;
	price: number;
	description: string;
  image: string | File;
}

interface Category {
  id: number;
  name: string;
}

export const Edit = () => {
  const navigate = useNavigate();
  const params = useParams();
  const [item, setItem] = useState<Item>();
  const [itemImage, setItemImage] = useState<Blob>();
  const [cookies] = useCookies(["token", "userID"]);

  const initialState = {
		id: params.id ? parseInt(params.id) : 0,
    name: "",
    category_id: 1,
    price: 0,
    description: "",
    image: "",
  };

  const fetchItem = () => {
    fetcher<Item>(`/items/${params.id}`, {
      method: "GET",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
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

  useEffect(() => {
    fetchItem();
  }, []);

  const [values, setValues] = useState<FormDataType>(initialState);
  const [file, setFile] = useState<string>(`${process.env.PUBLIC_URL}/default.jpg`);
  const [categories, setCategories] = useState<Category[]>([]);

  const onValueChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values,
      [event.target.name]: event.target.value,
    });
  };

  const onSelectChange = (event: React.ChangeEvent<HTMLSelectElement>) => {
    setValues({
      ...values,
      [event.target.name]: event.target.value,
    });
  };

  const onFileChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    setValues({
      ...values,
      [event.target.name]: event.target.files![0],
    });
		if(event.target.files![0])
			setFile(URL.createObjectURL(event.target.files![0]))
		else
			setFile(`${process.env.PUBLIC_URL}/default.jpg`)
  };

  const onSubmit = (event: React.FormEvent<HTMLFormElement>) => {
    event.preventDefault();
    const data = new FormData();
    data.append("name", values.name);
    data.append("category_id", values.category_id.toString());
    data.append("price", values.price.toString());
    data.append("description", values.description);
    data.append("image", values.image);

    fetcher<{ id: number }>(`/items`, {
      method: "POST",
      body: data,
      headers: {
        Authorization: `Bearer ${cookies.token}`,
      },
    })
      .then((res) => {
        sell(res.id);
      })
      .catch((error: Error) => {
        toast.error(error.message);
        console.error("POST error:", error);
      });
  };

  const sell = (itemID: number) =>
    fetcher(`/sell`, {
      method: "POST",
      headers: {
        Accept: "application/json",
        "Content-Type": "application/json",
        Authorization: `Bearer ${cookies.token}`,
      },
      body: JSON.stringify({
        item_id: itemID,
        user_id: parseInt(cookies.userID),
      }),
    })
      .then((_) => {
        toast.success("Item added successfully!");
      })
      .catch((error: Error) => {
        toast.error(error.message);
        console.error("POST error:", error);
      });

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

  useEffect(() => {
    fetchCategories();
  }, []);

  return (
    <MerComponent>
      <div className="Listing">
        <form onSubmit={onSubmit} className="ListingForm">
          <div>
            <img
              height={480}
              width={480}
              src={itemImage ? URL.createObjectURL(itemImage) : undefined}
              alt="item"
            />
            <span>Item Name(Before)</span>
            <input
              type="text"
              name="name"
              id="MerTextInput"
              placeholder="name"
              onChange={onValueChange}
							defaultValue={item?.name}
              required
            />
            <select
              name="category_id"
              id="MerTextInput"
              value={values.category_id}
              onChange={onSelectChange}
							defaultValue={item?.category_id}
            >
              {categories &&
                categories.map((category) => {
                  return <option value={category.id}>{category.name}</option>;
                })}
            </select>
            <input
              type="number"
              name="price"
              id="MerTextInput"
              placeholder="price"
              onChange={onValueChange}
							defaultValue={item?.price}
              required
            />
            <input
              type="text"
              name="description"
              id="MerTextInput"
              placeholder="description"
              onChange={onValueChange}
							defaultValue={item?.description}
              required
            />
            <input
              type="file"
              name="image"
              id="MerTextInput"
              onChange={onFileChange}
              required
            />
            <button type="submit" id="MerButton">
              List this item
            </button>
          </div>
        </form>
      </div>
    </MerComponent>
  );
};
