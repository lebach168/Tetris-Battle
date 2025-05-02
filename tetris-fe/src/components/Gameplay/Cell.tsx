"use client"

import React from "react";

type BlockProps = {
  type: string; // '0' (empty), '1', '2', '3', '4', 'ghost'
  size?: number;
};

const Block: React.FC<BlockProps> = ({ type, size = 20 }) => {
  const borderThickness = 3;
  const innerSize = size - borderThickness;

  // Base block với border đen cạnh dưới + phải
  const baseStyle = {
    width: `${size}px`,
    height: `${size}px`,
    boxSizing: "border-box" as const,
    position: "relative" as const,
    borderRight: `${borderThickness}px solid black`,
    borderBottom: `${borderThickness}px solid black`,
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
  };

  // Inner block (nội dung bên trong ô)
  const innerStyle = {
    width: `${innerSize}px`,
    height: `${innerSize}px`,
    position: "relative" as const,
    boxSizing: "border-box" as const,
  };

  if (type === "0") {
    return <div style={{ ...baseStyle, borderRight: "none", borderBottom: "none" }} />;
  }

  if (type === "ghost") {
    return (
      <div style={baseStyle}>
        <div
          style={{
            ...innerStyle,
            backgroundColor: "transparent",
            border: "2px dashed #888",
          }}
        />
      </div>
    );
  }

  const renderPixels = () => (
    <>
      <div
        style={{
          width: "2px",
          height: "2px",
          backgroundColor: "white",
          position: "absolute",
          top: "0px",
          left: "0px",
        }}
      />
      <div
        style={{
          width: "4px",
          height: "2px",
          backgroundColor: "white",
          position: "absolute",
          top: "2px",
          left: "2px",
        }}
      />
      <div
        style={{
          width: "2px",
          height: "4px",
          backgroundColor: "white",
          position: "absolute",
          top: "2px",
          left: "2px",
        }}
      />
    </>
  );

  let bg = "";
  let border = "";

  switch (type) {
    case "1":
      bg = "white";
      border = "2px solid #00CC66";
      break;
    case "2":
      bg = "#00CC66";
      break;
    case "3":
      bg = "#3366FF";
      break;
    case "4":
      bg = "#999999";//#BBBBBB  #777777  #666666
      break;
    default:
      return null;
  }

  return (
    <div style={baseStyle}>
      <div
        style={{
          ...innerStyle,
          backgroundColor: bg,
          border: border || "none",
        }}
      >
        {type !== "1" ? renderPixels() : (
          <div
            style={{
              width: "2px",
              height: "2px",
              backgroundColor: "white",
              position: "absolute",
              top: "0px",
              left: "0px",
            }}
          />
        )}
      </div>
    </div>
  );
};

export default React.memo(Block);
