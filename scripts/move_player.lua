unction onKeyPress(key)
    local moved, newX, newY = false, player.x, player.y

    if key == "right" then
        moved, newX, newY = attemptMove(player, 1, 0)
    elseif key == "left" then
        moved, newX, newY = attemptMove(player, -1, 0)
    end

    print("Moved:", moved, "New Position:", newX, newY)
end