import React from "react";
import { Routes, Route, BrowserRouter } from "react-router-dom";
import { Home } from "./components/Home";
import { ItemDetail } from "./components/ItemDetail";
import { Edit } from "./components/Edit";
import { UserProfile } from "./components/UserProfile";
import { Listing } from "./components/Listing";
import "./App.css";
import { Header } from "./components/Header/Header";
import { ToastContainer } from 'react-toastify';
import 'react-toastify/dist/ReactToastify.css';
import { PurchaseHistory } from "./components/PurchaseHisotry/PurchaseHistory";

export const App: React.VFC = () => {
  return (
    <>
      <ToastContainer position="bottom-center"/>

      <BrowserRouter>
        <div className="MerComponent">
          <Header></Header>
          <Routes>
            <Route index element={<Home />} />
            <Route path="/item/:id" element={<ItemDetail />} />
            <Route path="/item/:id/edit" element={<Edit />} />
            <Route path="/user/:id" element={<UserProfile />} />
            <Route path="/user/:id/purchase" element={<PurchaseHistory />} />
            <Route path="/sell" element={<Listing />} />
          </Routes>
        </div>
      </BrowserRouter>
    </>
  );
};
